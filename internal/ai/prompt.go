package ai

import (
	"encoding/json"
	"fmt"
	"sort"

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

Remember: Most network activity is normal. Only flag actual security issues, not routine internet usage.`

// BuildPrompt constructs the full prompt with system context
func BuildPrompt(userQuery string, sysCtx models.SystemContext) string {
	// Summarize system data for the prompt
	summary := summarizeSystemData(sysCtx)

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

func summarizeSystemData(ctx models.SystemContext) string {
	var summary string

	// CPU, Memory, and Disk summary
	summary += fmt.Sprintf("System Overview:\n- CPU Usage: %.1f%%\n- Memory Usage: %.1f%%\n- Disk Usage: %.1f%% (%v / %v GB)\n\n",
		ctx.CPUUsage, ctx.MemUsage, ctx.DiskUsage, ctx.DiskUsedGB, ctx.DiskTotalGB)

	// Top processes by CPU
	if len(ctx.Processes) > 0 {
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
			summary += fmt.Sprintf("- %s (PID %d): CPU %.1f%%, Memory %.1f MB\n",
				p.Name, p.PID, p.CPUPercent, p.MemoryMB)
		}
		summary += "\n"
	}

	// Open ports summary
	if len(ctx.Ports) > 0 {
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
				summary += fmt.Sprintf("- Port %d (%s): %s\n", p.Port, p.Protocol, p.ProcessName)
				count++
			}
		}
		summary += "\n"
	}

	// Security information including geo-located connections
	if ctx.SecurityInfo != nil {
		summary += "Security Information:\n"
		summary += fmt.Sprintf("- Firewall Status: %s\n", ctx.SecurityInfo.FirewallStatus)

		if len(ctx.SecurityInfo.SuspiciousProcs) > 0 {
			summary += fmt.Sprintf("- Suspicious Processes Detected: %d\n", len(ctx.SecurityInfo.SuspiciousProcs))
			for _, proc := range ctx.SecurityInfo.SuspiciousProcs {
				reasons := ""
				if len(proc.Reasons) > 0 {
					reasons = proc.Reasons[0] // Show first reason
					if len(proc.Reasons) > 1 {
						reasons += fmt.Sprintf(" (and %d more)", len(proc.Reasons)-1)
					}
				}
				summary += fmt.Sprintf("  * %s: %s\n", proc.Name, reasons)
			}
		}

		if len(ctx.SecurityInfo.UnknownConns) > 0 {
			summary += fmt.Sprintf("- External Network Connections: %d (normal internet activity)\n", len(ctx.SecurityInfo.UnknownConns))

			// Group connections by country
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

			// Show some specific connections with process context
			summary += "  Active Connections (by process):\n"
			count := 0
			for _, conn := range ctx.SecurityInfo.UnknownConns {
				if count < 6 {
					locationInfo := ""
					if conn.Country != "" {
						locationInfo = fmt.Sprintf(" [%s", conn.Country)
						if conn.City != "" {
							locationInfo += fmt.Sprintf(", %s", conn.City)
						}
						locationInfo += "]"
					}
					summary += fmt.Sprintf("    - %s -> %s%s (%s)\n",
						conn.LocalAddr, conn.RemoteAddr, locationInfo, conn.ProcessName)
					count++
				}
			}
		}
		summary += "\n"
	}

	// Network usage summary
	if len(ctx.Network) > 0 {
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
			summary += fmt.Sprintf("- %s: Download %.1f KB/s, Upload %.1f KB/s\n",
				n.ProcessName, n.DownloadSpeed/1024, n.UploadSpeed/1024)
		}
	}

	return summary
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
