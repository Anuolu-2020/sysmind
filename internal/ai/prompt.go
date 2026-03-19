package ai

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"sysmind/internal/models"
)

const systemPrompt = `You are SysMind, an AI-powered system monitoring assistant. Your role is to help users understand what their computer is doing in real-time in a clear, informative, and balanced way.

You analyze system data including:
- Running processes (CPU usage, memory usage)
- Disk usage and storage capacity
- Open network ports and connections
- Network bandwidth usage per process
- Security information including firewall status
- External network connections with geographic location (country, city)
- Process identification and normal vs suspicious activity

When responding:
1. Be clear, concise, and informative
2. Explain technical concepts in simple terms
3. Provide context for why connections exist (normal web browsing, app updates, etc.)
4. Focus on answering the user's specific question directly
5. Only mention security concerns for genuinely suspicious activity
6. Treat common ports (80/HTTP, 443/HTTPS) and well-known services as normal
7. When discussing connections, mention which processes are responsible
8. Use a helpful, educational tone rather than alarmist
9. IMPORTANT: Do NOT use markdown formatting like **bold** or *italic* - respond in plain text only
10. Use simple, readable formatting with line breaks and indentation, but no special markup

Remember: Most network activity is normal. Only flag actual security issues, not routine internet usage.`

// BuildPrompt constructs the full prompt with system context, respecting privacy settings
func BuildPrompt(userQuery string, sysCtx models.SystemContext, privacy models.PrivacyConfig) string {
	// Summarize system data for the prompt with privacy controls
	summary := summarizeSystemDataWithPrivacy(sysCtx, privacy)

	prompt := fmt.Sprintf(`Current System State:
%s

User Question: %s

Please provide a clear, informative response. Focus on directly answering the user's question. 

Context notes:
- Port 443 = HTTPS (secure web traffic) - completely normal
- Port 80 = HTTP (web traffic) - normal
- Popular services (Google, CDNs, social media) are expected
- Most network connections are routine internet usage
- Only mention security concerns for genuinely suspicious activity`, summary, userQuery)

	return prompt
}

// summarizeSystemDataWithPrivacy creates a summary respecting privacy settings
func summarizeSystemDataWithPrivacy(ctx models.SystemContext, privacy models.PrivacyConfig) string {
	var summary string

	// CPU, Memory, and Disk summary (basic stats)
	if privacy.ShareSystemStats {
		summary += fmt.Sprintf("System Overview:\n- CPU Usage: %.1f%%\n- Memory Usage: %.1f%%\n- Disk Usage: %.1f%% (%v / %v GB)\n\n",
			ctx.CPUUsage, ctx.MemUsage, ctx.DiskUsage, ctx.DiskUsedGB, ctx.DiskTotalGB)
	} else {
		summary += "System Overview: [Privacy Protected]\n\n"
	}

	// Top processes by CPU
	if len(ctx.Processes) > 0 && (privacy.ShareProcessNames || privacy.ShareProcessDetails) {
		procs := make([]models.ProcessInfo, len(ctx.Processes))
		copy(procs, ctx.Processes)
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].CPUPercent > procs[j].CPUPercent
		})

		summary += "Top Processes by CPU:\n"
		count := 10
		if len(procs) < count {
			count = len(procs)
		}
		for i := 0; i < count; i++ {
			p := procs[i]
			procName := p.Name
			if privacy.AnonymizeProcesses {
				procName = categorizeProcess(p.Name)
			}

			if privacy.ShareProcessNames && privacy.ShareProcessDetails {
				summary += fmt.Sprintf("- %s (PID %d): CPU %.1f%%, Memory %.1f MB\n",
					procName, p.PID, p.CPUPercent, p.MemoryMB)
			} else if privacy.ShareProcessNames {
				summary += fmt.Sprintf("- %s\n", procName)
			} else if privacy.ShareProcessDetails {
				summary += fmt.Sprintf("- [Process]: CPU %.1f%%, Memory %.1f MB\n",
					p.CPUPercent, p.MemoryMB)
			}
		}
		summary += "\n"
	}

	// Open ports summary
	if len(ctx.Ports) > 0 && privacy.ShareNetworkPorts {
		summary += "Open Ports:\n"
		listeningPorts := 0
		establishedConns := 0
		for _, p := range ctx.Ports {
			if p.State == "LISTENING" {
				listeningPorts++
			} else if p.State == "ESTABLISHED" {
				establishedConns++
			}
		}
		summary += fmt.Sprintf("- Listening ports: %d\n- Established connections: %d\n", listeningPorts, establishedConns)

		// Show some specific listening ports
		summary += "Notable listening ports:\n"
		count := 0
		for _, p := range ctx.Ports {
			if p.State == "LISTENING" && count < 10 {
				procName := p.ProcessName
				if privacy.AnonymizeProcesses {
					procName = categorizeProcess(p.ProcessName)
				}
				summary += fmt.Sprintf("- Port %d (%s): %s\n", p.Port, p.Protocol, procName)
				count++
			}
		}
		summary += "\n"
	}

	// Security information including geo-located connections
	if ctx.SecurityInfo != nil && privacy.ShareSecurityInfo {
		summary += "Security Information:\n"
		summary += fmt.Sprintf("- Firewall Status: %s\n", ctx.SecurityInfo.FirewallStatus)

		if len(ctx.SecurityInfo.SuspiciousProcs) > 0 {
			summary += fmt.Sprintf("- Suspicious Processes Detected: %d\n", len(ctx.SecurityInfo.SuspiciousProcs))
			for _, proc := range ctx.SecurityInfo.SuspiciousProcs {
				procName := proc.Name
				if privacy.AnonymizeProcesses {
					procName = categorizeProcess(proc.Name)
				}
				reasons := ""
				if len(proc.Reasons) > 0 {
					reasons = proc.Reasons[0] // Show first reason
					if len(proc.Reasons) > 1 {
						reasons += fmt.Sprintf(" (and %d more)", len(proc.Reasons)-1)
					}
				}
				summary += fmt.Sprintf("  * %s: %s\n", procName, reasons)
			}
		}

		if len(ctx.SecurityInfo.UnknownConns) > 0 && (privacy.ShareConnectionIPs || privacy.ShareConnectionGeo) {
			summary += fmt.Sprintf("- External Network Connections: %d (normal internet activity)\n", len(ctx.SecurityInfo.UnknownConns))

			// Group connections by country if geo sharing is enabled
			if privacy.ShareConnectionGeo {
				countryMap := make(map[string]int)
				for _, conn := range ctx.SecurityInfo.UnknownConns {
					if conn.Country != "" {
						countryMap[conn.Country]++
					}
				}

				if len(countryMap) > 0 {
					summary += "  Geographic Distribution:\n"
					for country, count := range countryMap {
						summary += fmt.Sprintf("    - %s: %d connection(s)\n", country, count)
					}
				}
			}

			// Show some specific connections with process context
			if privacy.ShareConnectionIPs {
				summary += "  Active Connections (by process):\n"
				count := 0
				for _, conn := range ctx.SecurityInfo.UnknownConns {
					if count < 6 {
						procName := conn.ProcessName
						if privacy.AnonymizeProcesses {
							procName = categorizeProcess(conn.ProcessName)
						}

						remoteAddr := conn.RemoteAddr
						if privacy.AnonymizeConnections {
							remoteAddr = categorizeIP(conn.RemoteAddr)
						}

						locationInfo := ""
						if privacy.ShareConnectionGeo && conn.Country != "" {
							locationInfo = fmt.Sprintf(" [%s", conn.Country)
							if conn.City != "" {
								locationInfo += fmt.Sprintf(", %s", conn.City)
							}
							locationInfo += "]"
						}
						summary += fmt.Sprintf("    - %s -> %s%s (%s)\n",
							conn.LocalAddr, remoteAddr, locationInfo, procName)
						count++
					}
				}
			}
		}
		summary += "\n"
	}

	// Network usage summary
	if len(ctx.Network) > 0 && privacy.ShareProcessDetails {
		netUsage := make([]models.NetworkUsage, len(ctx.Network))
		copy(netUsage, ctx.Network)
		sort.Slice(netUsage, func(i, j int) bool {
			return (netUsage[i].DownloadSpeed + netUsage[i].UploadSpeed) >
				(netUsage[j].DownloadSpeed + netUsage[j].UploadSpeed)
		})

		summary += "Top Network Usage:\n"
		count := 5
		if len(netUsage) < count {
			count = len(netUsage)
		}
		for i := 0; i < count; i++ {
			n := netUsage[i]
			procName := n.ProcessName
			if privacy.AnonymizeProcesses {
				procName = categorizeProcess(n.ProcessName)
			}
			summary += fmt.Sprintf("- %s: Download %.1f KB/s, Upload %.1f KB/s\n",
				procName, n.DownloadSpeed/1024, n.UploadSpeed/1024)
		}
	}

	return summary
}

// categorizeProcess replaces specific process names with categories
func categorizeProcess(name string) string {
	nameLower := strings.ToLower(name)

	// Browsers
	browsers := []string{"firefox", "chrome", "chromium", "safari", "edge", "opera", "brave", "vivaldi"}
	for _, b := range browsers {
		if strings.Contains(nameLower, b) {
			return "[Browser]"
		}
	}

	// Development tools
	devTools := []string{"code", "vscode", "vim", "nvim", "emacs", "idea", "pycharm", "webstorm", "node", "npm", "yarn", "go", "python", "java", "rust", "cargo"}
	for _, d := range devTools {
		if strings.Contains(nameLower, d) {
			return "[Dev Tool]"
		}
	}

	// Communication
	commApps := []string{"slack", "teams", "discord", "zoom", "skype", "telegram", "signal", "whatsapp"}
	for _, c := range commApps {
		if strings.Contains(nameLower, c) {
			return "[Communication]"
		}
	}

	// Media
	mediaApps := []string{"spotify", "vlc", "mpv", "music", "video", "player"}
	for _, m := range mediaApps {
		if strings.Contains(nameLower, m) {
			return "[Media]"
		}
	}

	// System processes
	sysProcs := []string{"systemd", "kernel", "init", "dbus", "udev", "journald", "networkmanager", "pulseaudio", "pipewire"}
	for _, s := range sysProcs {
		if strings.Contains(nameLower, s) {
			return "[System]"
		}
	}

	// Database
	dbProcs := []string{"postgres", "mysql", "mongo", "redis", "sqlite"}
	for _, db := range dbProcs {
		if strings.Contains(nameLower, db) {
			return "[Database]"
		}
	}

	return "[Application]"
}

// categorizeIP replaces specific IP addresses with provider categories
func categorizeIP(ip string) string {
	// Extract just the IP without port
	ipOnly := ip
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ipOnly = ip[:idx]
	}

	// Common CDN/Cloud ranges (simplified - in production you'd use proper IP range lookups)
	if strings.HasPrefix(ipOnly, "142.250.") || strings.HasPrefix(ipOnly, "172.217.") {
		return "[Google Services]"
	}
	if strings.HasPrefix(ipOnly, "52.") || strings.HasPrefix(ipOnly, "54.") || strings.HasPrefix(ipOnly, "3.") {
		return "[AWS Cloud]"
	}
	if strings.HasPrefix(ipOnly, "13.") || strings.HasPrefix(ipOnly, "20.") || strings.HasPrefix(ipOnly, "40.") {
		return "[Microsoft/Azure]"
	}
	if strings.HasPrefix(ipOnly, "104.") || strings.HasPrefix(ipOnly, "172.64.") || strings.HasPrefix(ipOnly, "1.1.") {
		return "[Cloudflare CDN]"
	}
	if strings.HasPrefix(ipOnly, "151.101.") || strings.HasPrefix(ipOnly, "199.232.") {
		return "[Fastly CDN]"
	}
	if strings.HasPrefix(ipOnly, "185.199.") {
		return "[GitHub]"
	}
	if strings.HasPrefix(ipOnly, "157.240.") || strings.HasPrefix(ipOnly, "31.13.") {
		return "[Meta/Facebook]"
	}

	return "[External Server]"
}

// GetSystemPrompt returns the system prompt for the AI
func GetSystemPrompt() string {
	return systemPrompt
}

// ParseAIResponse attempts to parse structured response from AI
func ParseAIResponse(response string) models.AIResponse {
	// Try to extract structured information from the response
	result := models.AIResponse{
		Explanation: response,
		RiskLevel:   "low",
		Suggestions: []string{},
	}

	// More balanced heuristics for risk level - only flag genuine threats
	if containsAny(response, []string{"malware", "virus", "trojan", "ransomware", "backdoor", "keylogger", "rootkit", "botnet"}) {
		result.RiskLevel = "high"
	} else if containsAny(response, []string{"unusual behavior", "unexpected process", "unknown executable", "suspicious activity", "investigate further"}) {
		result.RiskLevel = "medium"
	}
	// Note: Removed common terms like "suspicious", "concern", "monitor" that could flag normal activity

	return result
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ToJSON converts system context to JSON for debugging
func ToJSON(ctx models.SystemContext) string {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return string(data)
}
