package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GeoIPResult contains geolocation information for an IP address
type GeoIPResult struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	City        string  `json:"city"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
}

// GeoIPService provides IP geolocation lookup
type GeoIPService struct {
	cache      map[string]*GeoIPResult
	cacheMutex sync.RWMutex
	client     *http.Client
}

// NewGeoIPService creates a new GeoIP service instance
func NewGeoIPService() *GeoIPService {
	return &GeoIPService{
		cache: make(map[string]*GeoIPResult),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ClearCache clears the GeoIP cache
func (g *GeoIPService) ClearCache() {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	g.cache = make(map[string]*GeoIPResult)
}

// LookupIP performs a geolocation lookup for an IP address
// Uses ip-api.com free API (no API key required, 45 req/min limit)
func (g *GeoIPService) LookupIP(ipAddr string) (*GeoIPResult, error) {
	// Extract IP from address (remove port)
	ip := extractIP(ipAddr)
	if ip == "" {
		return nil, fmt.Errorf("invalid IP address: %s", ipAddr)
	}

	// Check if IP is private/local
	if isPrivateIP(ip) {
		return &GeoIPResult{
			Country:     "Local Network",
			CountryCode: "LAN",
			City:        "Local",
			Latitude:    0,
			Longitude:   0,
		}, nil
	}

	// Check cache first
	g.cacheMutex.RLock()
	if cached, exists := g.cache[ip]; exists {
		g.cacheMutex.RUnlock()
		return cached, nil
	}
	g.cacheMutex.RUnlock()

	// Make API request with English language
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,city,lat,lon&lang=en", ip)
	resp, err := g.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var apiResp struct {
		Status      string  `json:"status"`
		Message     string  `json:"message"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	result := &GeoIPResult{
		Country:     forceEnglishCountryName(apiResp.CountryCode),
		CountryCode: apiResp.CountryCode,
		City:        apiResp.City,
		Latitude:    apiResp.Lat,
		Longitude:   apiResp.Lon,
	}

	// Cache the result
	g.cacheMutex.Lock()
	g.cache[ip] = result
	g.cacheMutex.Unlock()

	return result, nil
}

// LookupBatch performs geolocation lookup for multiple IPs (with rate limiting)
func (g *GeoIPService) LookupBatch(ipAddrs []string) map[string]*GeoIPResult {
	results := make(map[string]*GeoIPResult)
	resultsMutex := sync.Mutex{}

	// Rate limit: max 45 requests per minute (ip-api.com free tier)
	// Add a small delay between requests to avoid hitting the limit
	rateLimiter := time.NewTicker(1500 * time.Millisecond) // ~40 req/min
	defer rateLimiter.Stop()

	for _, addr := range ipAddrs {
		ip := extractIP(addr)
		if ip == "" {
			continue
		}

		// Check cache first
		g.cacheMutex.RLock()
		if cached, exists := g.cache[ip]; exists {
			resultsMutex.Lock()
			results[ip] = cached
			resultsMutex.Unlock()
			g.cacheMutex.RUnlock()
			continue
		}
		g.cacheMutex.RUnlock()

		// Wait for rate limiter
		<-rateLimiter.C

		// Lookup IP
		result, err := g.LookupIP(addr)
		if err == nil {
			resultsMutex.Lock()
			results[ip] = result
			resultsMutex.Unlock()
		}
	}

	return results
}

// extractIP extracts the IP address from "ip:port" format
func extractIP(addr string) string {
	if addr == "" {
		return ""
	}

	// Handle IPv6 addresses with port [ip]:port
	if strings.HasPrefix(addr, "[") {
		end := strings.Index(addr, "]")
		if end > 0 {
			return addr[1:end]
		}
	}

	// Handle IPv4 addresses with port ip:port
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		return host
	}

	// No port, return as-is
	return addr
}

// isPrivateIP checks if an IP address is private/local
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// forceEnglishCountryName always uses our English mapping based on country code
func forceEnglishCountryName(countryCode string) string {
	if englishName, exists := countryCodeToEnglish[countryCode]; exists {
		return englishName
	}
	return countryCode // fallback to country code if not found
}

// containsNonLatin checks if a string contains non-Latin characters
func containsNonLatin(s string) bool {
	for _, r := range s {
		// Allow Latin characters, spaces, hyphens, apostrophes
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == ' ' || r == '-' || r == '\'') {
			return true
		}
	}
	return false
}

// countryCodeToEnglish maps ISO country codes to English names
var countryCodeToEnglish = map[string]string{
	"AD": "Andorra",
	"AE": "United Arab Emirates",
	"AF": "Afghanistan",
	"AG": "Antigua and Barbuda",
	"AI": "Anguilla",
	"AL": "Albania",
	"AM": "Armenia",
	"AO": "Angola",
	"AQ": "Antarctica",
	"AR": "Argentina",
	"AS": "American Samoa",
	"AT": "Austria",
	"AU": "Australia",
	"AW": "Aruba",
	"AX": "Åland Islands",
	"AZ": "Azerbaijan",
	"BA": "Bosnia and Herzegovina",
	"BB": "Barbados",
	"BD": "Bangladesh",
	"BE": "Belgium",
	"BF": "Burkina Faso",
	"BG": "Bulgaria",
	"BH": "Bahrain",
	"BI": "Burundi",
	"BJ": "Benin",
	"BL": "Saint Barthélemy",
	"BM": "Bermuda",
	"BN": "Brunei",
	"BO": "Bolivia",
	"BQ": "Bonaire",
	"BR": "Brazil",
	"BS": "Bahamas",
	"BT": "Bhutan",
	"BV": "Bouvet Island",
	"BW": "Botswana",
	"BY": "Belarus",
	"BZ": "Belize",
	"CA": "Canada",
	"CC": "Cocos Islands",
	"CD": "Democratic Republic of the Congo",
	"CF": "Central African Republic",
	"CG": "Republic of the Congo",
	"CH": "Switzerland",
	"CI": "Côte d'Ivoire",
	"CK": "Cook Islands",
	"CL": "Chile",
	"CM": "Cameroon",
	"CN": "China",
	"CO": "Colombia",
	"CR": "Costa Rica",
	"CU": "Cuba",
	"CV": "Cape Verde",
	"CW": "Curaçao",
	"CX": "Christmas Island",
	"CY": "Cyprus",
	"CZ": "Czech Republic",
	"DE": "Germany",
	"DJ": "Djibouti",
	"DK": "Denmark",
	"DM": "Dominica",
	"DO": "Dominican Republic",
	"DZ": "Algeria",
	"EC": "Ecuador",
	"EE": "Estonia",
	"EG": "Egypt",
	"EH": "Western Sahara",
	"ER": "Eritrea",
	"ES": "Spain",
	"ET": "Ethiopia",
	"FI": "Finland",
	"FJ": "Fiji",
	"FK": "Falkland Islands",
	"FM": "Micronesia",
	"FO": "Faroe Islands",
	"FR": "France",
	"GA": "Gabon",
	"GB": "United Kingdom",
	"GD": "Grenada",
	"GE": "Georgia",
	"GF": "French Guiana",
	"GG": "Guernsey",
	"GH": "Ghana",
	"GI": "Gibraltar",
	"GL": "Greenland",
	"GM": "Gambia",
	"GN": "Guinea",
	"GP": "Guadeloupe",
	"GQ": "Equatorial Guinea",
	"GR": "Greece",
	"GS": "South Georgia",
	"GT": "Guatemala",
	"GU": "Guam",
	"GW": "Guinea-Bissau",
	"GY": "Guyana",
	"HK": "Hong Kong",
	"HM": "Heard Island",
	"HN": "Honduras",
	"HR": "Croatia",
	"HT": "Haiti",
	"HU": "Hungary",
	"ID": "Indonesia",
	"IE": "Ireland",
	"IL": "Israel",
	"IM": "Isle of Man",
	"IN": "India",
	"IO": "British Indian Ocean Territory",
	"IQ": "Iraq",
	"IR": "Iran",
	"IS": "Iceland",
	"IT": "Italy",
	"JE": "Jersey",
	"JM": "Jamaica",
	"JO": "Jordan",
	"JP": "Japan",
	"KE": "Kenya",
	"KG": "Kyrgyzstan",
	"KH": "Cambodia",
	"KI": "Kiribati",
	"KM": "Comoros",
	"KN": "Saint Kitts and Nevis",
	"KP": "North Korea",
	"KR": "South Korea",
	"KW": "Kuwait",
	"KY": "Cayman Islands",
	"KZ": "Kazakhstan",
	"LA": "Laos",
	"LB": "Lebanon",
	"LC": "Saint Lucia",
	"LI": "Liechtenstein",
	"LK": "Sri Lanka",
	"LR": "Liberia",
	"LS": "Lesotho",
	"LT": "Lithuania",
	"LU": "Luxembourg",
	"LV": "Latvia",
	"LY": "Libya",
	"MA": "Morocco",
	"MC": "Monaco",
	"MD": "Moldova",
	"ME": "Montenegro",
	"MF": "Saint Martin",
	"MG": "Madagascar",
	"MH": "Marshall Islands",
	"MK": "North Macedonia",
	"ML": "Mali",
	"MM": "Myanmar",
	"MN": "Mongolia",
	"MO": "Macao",
	"MP": "Northern Mariana Islands",
	"MQ": "Martinique",
	"MR": "Mauritania",
	"MS": "Montserrat",
	"MT": "Malta",
	"MU": "Mauritius",
	"MV": "Maldives",
	"MW": "Malawi",
	"MX": "Mexico",
	"MY": "Malaysia",
	"MZ": "Mozambique",
	"NA": "Namibia",
	"NC": "New Caledonia",
	"NE": "Niger",
	"NF": "Norfolk Island",
	"NG": "Nigeria",
	"NI": "Nicaragua",
	"NL": "Netherlands",
	"NO": "Norway",
	"NP": "Nepal",
	"NR": "Nauru",
	"NU": "Niue",
	"NZ": "New Zealand",
	"OM": "Oman",
	"PA": "Panama",
	"PE": "Peru",
	"PF": "French Polynesia",
	"PG": "Papua New Guinea",
	"PH": "Philippines",
	"PK": "Pakistan",
	"PL": "Poland",
	"PM": "Saint Pierre and Miquelon",
	"PN": "Pitcairn Islands",
	"PR": "Puerto Rico",
	"PS": "Palestine",
	"PT": "Portugal",
	"PW": "Palau",
	"PY": "Paraguay",
	"QA": "Qatar",
	"RE": "Réunion",
	"RO": "Romania",
	"RS": "Serbia",
	"RU": "Russia",
	"RW": "Rwanda",
	"SA": "Saudi Arabia",
	"SB": "Solomon Islands",
	"SC": "Seychelles",
	"SD": "Sudan",
	"SE": "Sweden",
	"SG": "Singapore",
	"SH": "Saint Helena",
	"SI": "Slovenia",
	"SJ": "Svalbard and Jan Mayen",
	"SK": "Slovakia",
	"SL": "Sierra Leone",
	"SM": "San Marino",
	"SN": "Senegal",
	"SO": "Somalia",
	"SR": "Suriname",
	"SS": "South Sudan",
	"ST": "São Tomé and Príncipe",
	"SV": "El Salvador",
	"SX": "Sint Maarten",
	"SY": "Syria",
	"SZ": "Eswatini",
	"TC": "Turks and Caicos Islands",
	"TD": "Chad",
	"TF": "French Southern Territories",
	"TG": "Togo",
	"TH": "Thailand",
	"TJ": "Tajikistan",
	"TK": "Tokelau",
	"TL": "Timor-Leste",
	"TM": "Turkmenistan",
	"TN": "Tunisia",
	"TO": "Tonga",
	"TR": "Turkey",
	"TT": "Trinidad and Tobago",
	"TV": "Tuvalu",
	"TW": "Taiwan",
	"TZ": "Tanzania",
	"UA": "Ukraine",
	"UG": "Uganda",
	"UM": "United States Minor Outlying Islands",
	"US": "United States",
	"UY": "Uruguay",
	"UZ": "Uzbekistan",
	"VA": "Vatican City",
	"VC": "Saint Vincent and the Grenadines",
	"VE": "Venezuela",
	"VG": "British Virgin Islands",
	"VI": "United States Virgin Islands",
	"VN": "Vietnam",
	"VU": "Vanuatu",
	"WF": "Wallis and Futuna",
	"WS": "Samoa",
	"YE": "Yemen",
	"YT": "Mayotte",
	"ZA": "South Africa",
	"ZM": "Zambia",
	"ZW": "Zimbabwe",
}
