package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"sysmind/internal/ai"
	"sysmind/internal/collectors"
	"sysmind/internal/models"
	"sysmind/internal/services"
)

// App struct represents the main application
type App struct {
	ctx            context.Context
	collector      collectors.Collector
	configService  *services.ConfigService
	chatService    *services.ChatService
	geoIPService   *services.GeoIPService
	insightService *services.InsightService
	aiProvider     ai.Provider
	mu             sync.RWMutex

	// Cached system context
	cachedContext     models.SystemContext
	lastContextUpdate time.Time

	// Alerts
	alerts         []models.Alert
	alertConfig    models.AlertConfig
	alertMonitorOn bool
	alertTicker    *time.Ticker

	// Auto Insights
	insightMonitorOn bool
	insightTicker    *time.Ticker

	// Resource timeline history
	timelineHistory []models.ResourceTimelinePoint
	timelineMaxSize int
}

// NewApp creates a new App application struct
func NewApp() *App {
	configService, _ := services.NewConfigService()
	chatService, _ := services.NewChatService()
	geoIPService := services.NewGeoIPService()
	insightService, _ := services.NewInsightService()

	app := &App{
		collector:      collectors.NewCollector(),
		configService:  configService,
		chatService:    chatService,
		geoIPService:   geoIPService,
		insightService: insightService,
		alertConfig: models.AlertConfig{
			CPUThreshold:      80.0,
			MemoryThreshold:   85.0,
			DiskThreshold:     90.0,
			EnableAlerts:      true,
			EnableDesktopNotf: true,
			EnableSound:       false,
		},
		timelineHistory: make([]models.ResourceTimelinePoint, 0, 3600),
		timelineMaxSize: 3600,
	}

	// Initialize AI provider from config
	app.refreshAIProvider()

	return app
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Start alert monitoring
	a.startAlertMonitoring()

	// Start auto insights monitoring
	a.startInsightMonitoring()

	// Start timeline collection
	a.startTimelineMonitoring()
}

// startTimelineMonitoring starts a background collector for resource timeline data
func (a *App) startTimelineMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			stats, err := a.collector.GetDetailedStats()
			if err != nil || stats == nil {
				continue
			}

			point := models.ResourceTimelinePoint{
				Timestamp:      time.Now().UnixMilli(),
				CPUPercent:     stats.CPUPercent,
				MemoryPercent:  stats.MemoryPercent,
				DiskPercent:    stats.DiskPercent,
				NetUploadSpeed: stats.NetUploadSpeed,
				NetDownSpeed:   stats.NetDownSpeed,
			}

			a.mu.Lock()
			a.timelineHistory = append(a.timelineHistory, point)
			if len(a.timelineHistory) > a.timelineMaxSize {
				a.timelineHistory = a.timelineHistory[len(a.timelineHistory)-a.timelineMaxSize:]
			}
			a.mu.Unlock()
		}
	}()
}

// GetResourceTimeline returns recent timeline points for resource usage
func (a *App) GetResourceTimeline(minutes int) []models.ResourceTimelinePoint {
	if minutes <= 0 {
		minutes = 30
	}
	if minutes > 120 {
		minutes = 120
	}

	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute).UnixMilli()

	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]models.ResourceTimelinePoint, 0, len(a.timelineHistory))
	for _, point := range a.timelineHistory {
		if point.Timestamp >= cutoff {
			result = append(result, point)
		}
	}

	return result
}

// refreshAIProvider updates the AI provider based on current config
func (a *App) refreshAIProvider() {
	a.mu.Lock()
	defer a.mu.Unlock()

	config := a.configService.GetConfig()
	a.aiProvider = ai.NewProvider(config)
}

// GetProcesses returns list of running processes
func (a *App) GetProcesses() []models.ProcessInfo {
	procs, err := a.collector.GetProcesses()
	if err != nil {
		return []models.ProcessInfo{}
	}
	return procs
}

// GetPorts returns list of open ports
func (a *App) GetPorts() []models.PortInfo {
	ports, err := a.collector.GetPorts()
	if err != nil {
		return []models.PortInfo{}
	}
	return ports
}

// GetNetworkUsage returns network usage per process
func (a *App) GetNetworkUsage() []models.NetworkUsage {
	usage, err := a.collector.GetNetworkUsage()
	if err != nil {
		return []models.NetworkUsage{}
	}
	return usage
}

// GetSystemStats returns CPU and memory usage
func (a *App) GetSystemStats() map[string]float64 {
	cpu, mem, err := a.collector.GetSystemStats()
	if err != nil {
		return map[string]float64{"cpu": 0, "memory": 0}
	}
	return map[string]float64{"cpu": cpu, "memory": mem}
}

// GetSystemContext returns the full system context
func (a *App) GetSystemContext() models.SystemContext {
	procs, _ := a.collector.GetProcesses()
	ports, _ := a.collector.GetPorts()
	network, _ := a.collector.GetNetworkUsage()
	cpu, mem, _ := a.collector.GetSystemStats()

	// Include security information with geo-located connections
	securityInfo, _ := a.collector.GetSecurityInfo()
	if securityInfo != nil && len(securityInfo.UnknownConns) > 0 {
		// Enrich connections with geo data for AI analysis
		a.enrichConnectionsWithGeo(securityInfo.UnknownConns)
	}

	ctx := models.SystemContext{
		Processes:    procs,
		Ports:        ports,
		Network:      network,
		CPUUsage:     cpu,
		MemUsage:     mem,
		Timestamp:    time.Now(),
		SecurityInfo: securityInfo, // Add security info to context
	}

	a.mu.Lock()
	a.cachedContext = ctx
	a.lastContextUpdate = time.Now()
	a.mu.Unlock()

	return ctx
}

// AskAI sends a question to the AI and returns the response (legacy method)
func (a *App) AskAI(question string) map[string]interface{} {
	a.mu.RLock()
	provider := a.aiProvider
	cachedCtx := a.cachedContext
	lastUpdate := a.lastContextUpdate
	a.mu.RUnlock()

	// Refresh system context if it's older than 5 seconds
	if time.Since(lastUpdate) > 5*time.Second {
		cachedCtx = a.GetSystemContext()
	}

	if provider == nil || !provider.Available() {
		return map[string]interface{}{
			"success": false,
			"error":   "AI provider not configured. Please configure your API key in Settings.",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.GenerateResponse(ctx, question, cachedCtx)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":  true,
		"response": response,
	}
}

// === Contextual AI Explanation Methods ===

// ExplainProcess provides AI-powered explanation for a specific process
func (a *App) ExplainProcess(pid int32) map[string]interface{} {
	a.mu.RLock()
	provider := a.aiProvider
	a.mu.RUnlock()

	if provider == nil || !provider.Available() {
		return map[string]interface{}{
			"success": false,
			"error":   "AI provider not configured. Please configure your API key in Settings.",
		}
	}

	// Get process details
	processes, err := a.collector.GetProcesses()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to get process information",
		}
	}

	var targetProcess *models.ProcessInfo
	for _, proc := range processes {
		if proc.PID == pid {
			targetProcess = &proc
			break
		}
	}

	if targetProcess == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Process not found",
		}
	}

	// Create focused context for this process
	question := fmt.Sprintf(`Please explain this process in detail:

Process Name: %s
PID: %d
CPU Usage: %.1f%%
Memory Usage: %.1f MB
Status: %s

Please provide:
1. What this process does and its purpose
2. Whether this CPU/memory usage is normal for this process
3. If there are any concerns or recommendations
4. Common reasons why this process might be using high resources
5. Whether this process is safe/legitimate

Be concise but informative.`,
		targetProcess.Name, targetProcess.PID, targetProcess.CPUPercent,
		targetProcess.MemoryMB, targetProcess.Status)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get lightweight system context
	systemCtx := a.GetSystemContext()

	response, err := provider.GenerateResponse(ctx, question, systemCtx)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":  true,
		"response": response,
		"process":  targetProcess,
	}
}

// ExplainPort provides AI-powered explanation for a specific port
func (a *App) ExplainPort(port uint32, protocol string) map[string]interface{} {
	a.mu.RLock()
	provider := a.aiProvider
	a.mu.RUnlock()

	if provider == nil || !provider.Available() {
		return map[string]interface{}{
			"success": false,
			"error":   "AI provider not configured. Please configure your API key in Settings.",
		}
	}

	// Get port details
	ports, err := a.collector.GetPorts()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to get port information",
		}
	}

	var targetPort *models.PortInfo
	for _, p := range ports {
		if p.Port == port && p.Protocol == protocol {
			targetPort = &p
			break
		}
	}

	if targetPort == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Port not found",
		}
	}

	question := fmt.Sprintf(`Please explain this network port in detail:

Port: %d
Protocol: %s
Process: %s (PID: %d)
State: %s
Local Address: %s

Please provide:
1. What this port is commonly used for
2. Which applications typically use this port
3. Whether having this port open is normal/safe
4. Any security considerations or risks
5. Whether this port should be exposed or blocked

Be concise but informative.`,
		targetPort.Port, targetPort.Protocol, targetPort.ProcessName,
		targetPort.PID, targetPort.State, targetPort.LocalAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get lightweight system context
	systemCtx := a.GetSystemContext()

	response, err := provider.GenerateResponse(ctx, question, systemCtx)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":  true,
		"response": response,
		"port":     targetPort,
	}
}

// ExplainNetworkActivity provides AI-powered explanation for network patterns
func (a *App) ExplainNetworkActivity() map[string]interface{} {
	a.mu.RLock()
	provider := a.aiProvider
	a.mu.RUnlock()

	if provider == nil || !provider.Available() {
		return map[string]interface{}{
			"success": false,
			"error":   "AI provider not configured. Please configure your API key in Settings.",
		}
	}

	// Get network details
	networkUsage, err := a.collector.GetNetworkUsage()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to get network information",
		}
	}

	securityInfo, _ := a.collector.GetSecurityInfo()

	question := `Please analyze the current network activity and explain:

1. Which processes are using the most bandwidth
2. Whether the network usage patterns are normal
3. Any suspicious or concerning connections
4. Which applications might be responsible for high data usage
5. Recommendations for optimizing network performance

Focus on practical insights and actionable recommendations.`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create focused context
	systemCtx := models.SystemContext{
		Network:      networkUsage,
		SecurityInfo: securityInfo,
		Timestamp:    time.Now(),
	}

	response, err := provider.GenerateResponse(ctx, question, systemCtx)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":      true,
		"response":     response,
		"networkUsage": networkUsage,
		"securityInfo": securityInfo,
	}
}

// ExplainConnection provides AI analysis for a specific network connection
func (a *App) ExplainConnection(remoteAddr string, remotePort int) map[string]interface{} {
	a.mu.RLock()
	provider := a.aiProvider
	a.mu.RUnlock()

	if provider == nil || !provider.Available() {
		return map[string]interface{}{
			"success": false,
			"error":   "AI provider not configured. Please configure your API key in Settings.",
		}
	}

	// Get security info to find the connection
	securityInfo, err := a.collector.GetSecurityInfo()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to get security information",
		}
	}

	var targetConn *models.ConnectionInfo
	for _, conn := range securityInfo.UnknownConns {
		if conn.RemoteAddr == remoteAddr {
			targetConn = &conn
			break
		}
	}

	if targetConn == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Connection not found",
		}
	}

	// Enrich with geo data if available
	if targetConn.Country == "" {
		geoData, err := a.geoIPService.LookupIP(remoteAddr)
		if err == nil && geoData != nil {
			targetConn.Country = geoData.Country
			targetConn.City = geoData.City
		}
	}

	question := fmt.Sprintf(`Please analyze this network connection:

Remote Address: %s
Remote Port: %d
Country: %s
City: %s

Please provide:
1. What this connection likely represents (service/company/purpose)
2. Whether this connection is normal or suspicious
3. What type of application might be making this connection
4. Any security considerations or risks
5. Whether this connection should be allowed or blocked

Be specific about the IP address and location if recognizable.`,
		targetConn.RemoteAddr, remotePort, targetConn.Country, targetConn.City)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create minimal context
	systemCtx := models.SystemContext{
		SecurityInfo: securityInfo,
		Timestamp:    time.Now(),
	}

	response, err := provider.GenerateResponse(ctx, question, systemCtx)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":    true,
		"response":   response,
		"connection": targetConn,
	}
}

// SendChatMessage sends a message in a chat session and gets AI response
func (a *App) SendChatMessage(sessionID, question string) map[string]interface{} {
	// Add user message to session
	userMsg := a.chatService.AddMessage(sessionID, "user", question, "")
	if userMsg == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Session not found",
		}
	}

	a.mu.RLock()
	provider := a.aiProvider
	cachedCtx := a.cachedContext
	lastUpdate := a.lastContextUpdate
	a.mu.RUnlock()

	// Refresh system context if it's older than 5 seconds
	if time.Since(lastUpdate) > 5*time.Second {
		cachedCtx = a.GetSystemContext()
	}

	if provider == nil || !provider.Available() {
		errMsg := a.chatService.AddMessage(sessionID, "error", "AI provider not configured. Please configure your API key in Settings.", "")
		return map[string]interface{}{
			"success":     false,
			"error":       "AI provider not configured",
			"userMessage": userMsg,
			"aiMessage":   errMsg,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := provider.GenerateResponse(ctx, question, cachedCtx)
	if err != nil {
		errMsg := a.chatService.AddMessage(sessionID, "error", err.Error(), "")
		return map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"userMessage": userMsg,
			"aiMessage":   errMsg,
		}
	}

	parsed := ai.ParseAIResponse(response)
	aiMsg := a.chatService.AddMessage(sessionID, "assistant", response, parsed.RiskLevel)

	return map[string]interface{}{
		"success":     true,
		"userMessage": userMsg,
		"aiMessage":   aiMsg,
		"riskLevel":   parsed.RiskLevel,
	}
}

// SendChatMessageStreaming sends a message and streams the AI response
func (a *App) SendChatMessageStreaming(sessionID, question string) map[string]interface{} {
	// Add user message to session
	userMsg := a.chatService.AddMessage(sessionID, "user", question, "")
	if userMsg == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Session not found",
		}
	}

	a.mu.RLock()
	provider := a.aiProvider
	cachedCtx := a.cachedContext
	lastUpdate := a.lastContextUpdate
	a.mu.RUnlock()

	// Refresh system context if it's older than 5 seconds
	if time.Since(lastUpdate) > 5*time.Second {
		cachedCtx = a.GetSystemContext()
	}

	if provider == nil || !provider.Available() {
		errMsg := a.chatService.AddMessage(sessionID, "error", "AI provider not configured. Please configure your API key in Settings.", "")
		return map[string]interface{}{
			"success":     false,
			"error":       "AI provider not configured",
			"userMessage": userMsg,
			"aiMessage":   errMsg,
		}
	}

	// Generate response in background and emit chunks
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Emit start event
		runtime.EventsEmit(a.ctx, "chat:stream:start", map[string]interface{}{
			"sessionID": sessionID,
			"messageID": userMsg.ID,
		})

		response, err := provider.GenerateResponse(ctx, question, cachedCtx)
		if err != nil {
			runtime.EventsEmit(a.ctx, "chat:stream:error", map[string]interface{}{
				"sessionID": sessionID,
				"error":     err.Error(),
			})
			a.chatService.AddMessage(sessionID, "error", err.Error(), "")
			return
		}

		// Simulate streaming by chunking the response
		chunkSize := 20 // words per chunk
		words := splitIntoWords(response)
		fullText := ""

		for i := 0; i < len(words); i += chunkSize {
			end := i + chunkSize
			if end > len(words) {
				end = len(words)
			}
			chunk := joinWords(words[i:end])
			fullText += chunk + " "

			// Emit chunk
			runtime.EventsEmit(a.ctx, "chat:stream:chunk", map[string]interface{}{
				"sessionID": sessionID,
				"chunk":     chunk,
				"fullText":  fullText,
			})

			// Small delay to simulate streaming
			time.Sleep(50 * time.Millisecond)
		}

		// Save complete message
		parsed := ai.ParseAIResponse(response)
		aiMsg := a.chatService.AddMessage(sessionID, "assistant", response, parsed.RiskLevel)

		// Emit complete event
		runtime.EventsEmit(a.ctx, "chat:stream:complete", map[string]interface{}{
			"sessionID": sessionID,
			"message":   aiMsg,
			"riskLevel": parsed.RiskLevel,
		})
	}()

	return map[string]interface{}{
		"success":     true,
		"userMessage": userMsg,
		"streaming":   true,
	}
}

// Helper functions for word splitting
func splitIntoWords(text string) []string {
	words := []string{}
	currentWord := ""
	for _, char := range text {
		if char == ' ' || char == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
			if char == '\n' {
				words = append(words, "\n")
			}
		} else {
			currentWord += string(char)
		}
	}
	if currentWord != "" {
		words = append(words, currentWord)
	}
	return words
}

func joinWords(words []string) string {
	result := ""
	for i, word := range words {
		if word == "\n" {
			result += "\n"
		} else {
			result += word
			if i < len(words)-1 && words[i+1] != "\n" {
				result += " "
			}
		}
	}
	return result
}

// === Chat Session Methods ===

// CreateChatSession creates a new chat session
func (a *App) CreateChatSession(title string) *models.ChatSession {
	return a.chatService.CreateSession(title)
}

// GetChatSession retrieves a chat session by ID
func (a *App) GetChatSession(sessionID string) *models.ChatSession {
	return a.chatService.GetSession(sessionID)
}

// GetAllChatSessions returns all chat sessions
func (a *App) GetAllChatSessions() []models.ChatSessionSummary {
	return a.chatService.GetAllSessions()
}

// DeleteChatSession deletes a chat session
func (a *App) DeleteChatSession(sessionID string) error {
	return a.chatService.DeleteSession(sessionID)
}

// UpdateChatSessionTitle updates a session's title
func (a *App) UpdateChatSessionTitle(sessionID, title string) error {
	return a.chatService.UpdateSessionTitle(sessionID, title)
}

// ClearChatSession clears all messages in a session
func (a *App) ClearChatSession(sessionID string) error {
	return a.chatService.ClearSessionMessages(sessionID)
}

// === Config Methods ===

// GetAIConfig returns the current AI configuration
func (a *App) GetAIConfig() models.AIConfig {
	return a.configService.GetConfig()
}

// SetAIConfig updates the AI configuration
func (a *App) SetAIConfig(config models.AIConfig) error {
	err := a.configService.SetConfig(config)
	if err != nil {
		return err
	}
	a.refreshAIProvider()
	return nil
}

// GetAvailableProviders returns available AI providers and models
func (a *App) GetAvailableProviders() []services.ProviderInfo {
	return a.configService.GetAvailableProviders()
}

// IsAIConfigured checks if AI is properly configured
func (a *App) IsAIConfigured() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.aiProvider != nil && a.aiProvider.Available()
}

// === Enhanced System Monitoring Methods ===

// GetDetailedStats returns detailed system statistics
func (a *App) GetDetailedStats() *models.SystemStats {
	stats, err := a.collector.GetDetailedStats()
	if err != nil {
		return &models.SystemStats{}
	}
	return stats
}

// GetSecurityInfo returns security-related information
func (a *App) GetSecurityInfo() *models.SecurityInfo {
	info, err := a.collector.GetSecurityInfo()
	if err != nil {
		return &models.SecurityInfo{}
	}
	return info
}

// GetSecurityInfoWithGeo gets security info and enriches with geo data (may be slower)
func (a *App) GetSecurityInfoWithGeo() *models.SecurityInfo {
	info, err := a.collector.GetSecurityInfo()
	if err != nil || info == nil {
		return &models.SecurityInfo{}
	}

	// Enrich connections with geo data in background
	if len(info.UnknownConns) > 0 {
		a.enrichConnectionsWithGeoAsync(info.UnknownConns)
	}

	return info
}

func (a *App) GetDevEnvironmentInfo() *models.DevEnvironmentInfo {
	info, err := a.collector.GetDevEnvironmentInfo()
	if err != nil || info == nil {
		return &models.DevEnvironmentInfo{
			Containers:   []models.DockerContainer{},
			Environments: []models.DevEnvironment{},
			DevPorts:     []models.DevPort{},
		}
	}
	return info
}

// StartContainer starts a Docker container
func (a *App) StartContainer(containerID string) error {
	dockerService := services.NewDockerService()
	return dockerService.StartContainer(containerID)
}

// StopContainer stops a Docker container
func (a *App) StopContainer(containerID string) error {
	dockerService := services.NewDockerService()
	return dockerService.StopContainer(containerID)
}

// RestartContainer restarts a Docker container
func (a *App) RestartContainer(containerID string) error {
	dockerService := services.NewDockerService()
	return dockerService.RestartContainer(containerID)
}

// RemoveContainer removes a Docker container
func (a *App) RemoveContainer(containerID string) error {
	dockerService := services.NewDockerService()
	return dockerService.RemoveContainer(containerID)
}

// enrichConnectionsWithGeo adds geolocation data to connections (synchronous)
func (a *App) enrichConnectionsWithGeo(conns []models.ConnectionInfo) {
	// Limit to max 10 connections to prevent long delays
	maxConns := len(conns)
	if maxConns > 10 {
		maxConns = 10
	}

	for i := 0; i < maxConns; i++ {
		geoData, err := a.geoIPService.LookupIP(conns[i].RemoteAddr)
		if err == nil && geoData != nil {
			conns[i].Country = geoData.Country
			conns[i].CountryCode = geoData.CountryCode
			conns[i].City = geoData.City
			conns[i].Latitude = geoData.Latitude
			conns[i].Longitude = geoData.Longitude
		}
	}
}

// enrichConnectionsWithGeoAsync enriches connections in background
func (a *App) enrichConnectionsWithGeoAsync(conns []models.ConnectionInfo) {
	// Limit to max 5 connections to avoid too many API calls
	maxConns := len(conns)
	if maxConns > 5 {
		maxConns = 5
	}

	for i := 0; i < maxConns; i++ {
		geoData, err := a.geoIPService.LookupIP(conns[i].RemoteAddr)
		if err == nil && geoData != nil {
			// Note: This won't update the current response, but will be cached for next call
			// This is acceptable since Security tab refreshes every 10s
		}
	}
}

// KillProcess terminates a process by PID
func (a *App) KillProcess(pid int32) map[string]interface{} {
	err := a.collector.KillProcess(pid)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Process %d terminated", pid),
	}
}

// SetProcessPriority changes the priority of a process
func (a *App) SetProcessPriority(pid int32, priority int) map[string]interface{} {
	err := a.collector.SetProcessPriority(pid, priority)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Process %d priority set to %d", pid, priority),
	}
}

// === Alert Methods ===

// GetAlertConfig returns the current alert configuration
func (a *App) GetAlertConfig() models.AlertConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.alertConfig
}

// SetAlertConfig updates the alert configuration
func (a *App) SetAlertConfig(config models.AlertConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.alertConfig = config
}

// GetAlerts returns all active alerts
func (a *App) GetAlerts() []models.Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.alerts
}

// DismissAlert marks an alert as dismissed
func (a *App) DismissAlert(alertID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.alerts {
		if a.alerts[i].ID == alertID {
			a.alerts[i].Dismissed = true
			break
		}
	}
}

// ClearAlerts removes all dismissed alerts
func (a *App) ClearAlerts() {
	a.mu.Lock()
	defer a.mu.Unlock()
	var active []models.Alert
	for _, alert := range a.alerts {
		if !alert.Dismissed {
			active = append(active, alert)
		}
	}
	a.alerts = active
}

// === Chat Export Methods ===

// ExportChatSession exports a chat session to JSON or Markdown format
func (a *App) ExportChatSession(sessionID string, format string) map[string]interface{} {
	session := a.chatService.GetSession(sessionID)
	if session == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Session not found",
		}
	}

	var content string
	var filename string

	if format == "markdown" {
		content = a.exportToMarkdown(session)
		filename = fmt.Sprintf("sysmind-chat-%s.md", session.ID[:8])
	} else {
		data, err := json.MarshalIndent(session, "", "  ")
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		}
		content = string(data)
		filename = fmt.Sprintf("sysmind-chat-%s.json", session.ID[:8])
	}

	return map[string]interface{}{
		"success":  true,
		"content":  content,
		"filename": filename,
	}
}

func (a *App) exportToMarkdown(session *models.ChatSession) string {
	md := fmt.Sprintf("# %s\n\n", session.Title)
	md += fmt.Sprintf("*Exported from SysMind on %s*\n\n---\n\n", time.Now().Format("2006-01-02 15:04:05"))

	for _, msg := range session.Messages {
		timestamp := time.UnixMilli(msg.Timestamp).Format("15:04:05")
		switch msg.Role {
		case "user":
			md += fmt.Sprintf("## User (%s)\n\n%s\n\n", timestamp, msg.Content)
		case "assistant":
			md += fmt.Sprintf("## Assistant (%s)\n\n%s\n\n", timestamp, msg.Content)
		case "error":
			md += fmt.Sprintf("## Error (%s)\n\n> %s\n\n", timestamp, msg.Content)
		}
	}

	return md
}

// === Prompt Templates ===

// GetPromptTemplates returns predefined prompt templates
func (a *App) GetPromptTemplates() []models.PromptTemplate {
	return []models.PromptTemplate{
		{
			ID:          "security-audit",
			Name:        "Security Audit",
			Description: "Analyze system for security issues",
			Prompt:      "Perform a comprehensive security audit of my system. Check for suspicious processes, unusual network connections, and potential vulnerabilities.",
			Category:    "security",
			Icon:        "shield",
		},
		{
			ID:          "high-cpu",
			Name:        "High CPU Analysis",
			Description: "Find CPU-heavy processes",
			Prompt:      "Which processes are using the most CPU right now? Are any of them suspicious or unnecessary?",
			Category:    "performance",
			Icon:        "cpu",
		},
		{
			ID:          "high-memory",
			Name:        "Memory Usage",
			Description: "Analyze memory consumption",
			Prompt:      "Analyze memory usage on my system. Which processes are consuming the most memory? Any memory leaks?",
			Category:    "performance",
			Icon:        "memory",
		},
		{
			ID:          "network-analysis",
			Name:        "Network Analysis",
			Description: "Analyze network connections",
			Prompt:      "Show me all active network connections. Are there any unusual outbound connections or suspicious ports?",
			Category:    "network",
			Icon:        "network",
		},
		{
			ID:          "open-ports",
			Name:        "Open Ports Check",
			Description: "Review listening ports",
			Prompt:      "List all open ports on my system. Which services are listening and are any of them unnecessary or risky?",
			Category:    "network",
			Icon:        "port",
		},
		{
			ID:          "system-health",
			Name:        "System Health",
			Description: "Overall system health check",
			Prompt:      "Give me an overall health check of my system including CPU, memory, disk usage, and any potential issues.",
			Category:    "general",
			Icon:        "heart",
		},
		{
			ID:          "startup-apps",
			Name:        "Startup Analysis",
			Description: "Check startup applications",
			Prompt:      "What processes are currently running? Are there any that look like they shouldn't be running or could be startup bloat?",
			Category:    "performance",
			Icon:        "rocket",
		},
		{
			ID:          "root-cause-lag",
			Name:        "Why Is My PC Slow?",
			Description: "Root cause analysis for performance issues",
			Prompt:      "My PC has been running slowly/lagging. Can you analyze my current system state and identify the most likely causes? Look at CPU usage, memory consumption, disk activity, network traffic, and running processes. Correlate any spikes or high usage patterns and explain what might be causing the performance issues.",
			Category:    "performance",
			Icon:        "search",
		},
		{
			ID:          "bandwidth-hogs",
			Name:        "Bandwidth Usage",
			Description: "Find bandwidth-heavy apps",
			Prompt:      "Which applications are using the most network bandwidth right now? Are any of them uploading data unexpectedly?",
			Category:    "network",
			Icon:        "download",
		},
	}
}

// === Alert Monitoring ===

// GetQuickStats returns a quick summary for system tray/menu
func (a *App) GetQuickStats() map[string]interface{} {
	stats, err := a.collector.GetDetailedStats()
	if err != nil {
		return map[string]interface{}{
			"cpu":    0.0,
			"memory": 0.0,
			"disk":   0.0,
			"alerts": 0,
		}
	}

	a.mu.RLock()
	alertCount := len(a.alerts)
	for _, alert := range a.alerts {
		if alert.Dismissed {
			alertCount--
		}
	}
	a.mu.RUnlock()

	return map[string]interface{}{
		"cpu":    stats.CPUPercent,
		"memory": stats.MemoryPercent,
		"disk":   stats.DiskPercent,
		"alerts": alertCount,
	}
}

// startAlertMonitoring starts background monitoring for system alerts
func (a *App) startAlertMonitoring() {
	a.mu.Lock()
	if a.alertMonitorOn {
		a.mu.Unlock()
		return
	}
	a.alertMonitorOn = true
	a.mu.Unlock()

	// Check every 5 seconds
	a.alertTicker = time.NewTicker(5 * time.Second)

	go func() {
		for range a.alertTicker.C {
			a.checkForAlerts()
		}
	}()
}

// stopAlertMonitoring stops the alert monitoring
func (a *App) stopAlertMonitoring() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.alertTicker != nil {
		a.alertTicker.Stop()
		a.alertMonitorOn = false
	}
}

// checkForAlerts checks system stats against thresholds and creates alerts
func (a *App) checkForAlerts() {
	a.mu.RLock()
	config := a.alertConfig
	a.mu.RUnlock()

	if !config.EnableAlerts {
		return
	}

	stats, err := a.collector.GetDetailedStats()
	if err != nil {
		return
	}

	// Check CPU
	if stats.CPUPercent > config.CPUThreshold {
		a.addAlert(models.Alert{
			ID:        fmt.Sprintf("cpu-%d", time.Now().Unix()),
			Type:      "cpu",
			Severity:  "warning",
			Title:     "High CPU Usage",
			Message:   fmt.Sprintf("CPU usage is at %.1f%% (threshold: %.1f%%)", stats.CPUPercent, config.CPUThreshold),
			Timestamp: time.Now().UnixMilli(),
			Data: map[string]interface{}{
				"value":     stats.CPUPercent,
				"threshold": config.CPUThreshold,
			},
		})
	}

	// Check Memory
	if stats.MemoryPercent > config.MemoryThreshold {
		a.addAlert(models.Alert{
			ID:        fmt.Sprintf("memory-%d", time.Now().Unix()),
			Type:      "memory",
			Severity:  "warning",
			Title:     "High Memory Usage",
			Message:   fmt.Sprintf("Memory usage is at %.1f%% (threshold: %.1f%%)", stats.MemoryPercent, config.MemoryThreshold),
			Timestamp: time.Now().UnixMilli(),
			Data: map[string]interface{}{
				"value":     stats.MemoryPercent,
				"threshold": config.MemoryThreshold,
			},
		})
	}

	// Check Disk
	if stats.DiskPercent > config.DiskThreshold {
		a.addAlert(models.Alert{
			ID:        fmt.Sprintf("disk-%d", time.Now().Unix()),
			Type:      "disk",
			Severity:  "critical",
			Title:     "High Disk Usage",
			Message:   fmt.Sprintf("Disk usage is at %.1f%% (threshold: %.1f%%)", stats.DiskPercent, config.DiskThreshold),
			Timestamp: time.Now().UnixMilli(),
			Data: map[string]interface{}{
				"value":     stats.DiskPercent,
				"threshold": config.DiskThreshold,
			},
		})
	}
}

// addAlert adds a new alert and sends notification
func (a *App) addAlert(alert models.Alert) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if this alert already exists (avoid duplicates within 1 minute)
	now := time.Now().UnixMilli()
	for _, existing := range a.alerts {
		if existing.Type == alert.Type && !existing.Dismissed {
			// If there's a recent undismissed alert of the same type, skip
			if now-existing.Timestamp < 60000 { // 60 seconds
				return
			}
		}
	}

	a.alerts = append(a.alerts, alert)

	// Emit new alert event so frontend can notify regardless of active tab
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "alerts:new", map[string]interface{}{
			"alert":             alert,
			"enableDesktopNotf": a.alertConfig.EnableDesktopNotf,
			"enableSound":       a.alertConfig.EnableSound,
		})
	}

	// Send desktop notification if enabled
	if a.alertConfig.EnableDesktopNotf {
		a.sendNotification(alert.Title, alert.Message)
	}
}

// sendNotification sends a desktop notification
func (a *App) sendNotification(title, message string) {
	if a.ctx == nil {
		return
	}

	// Emit fallback event for frontend-side notifications.
	// This works cross-platform in the webview and does not depend on
	// a specific native notification API behavior.
	runtime.EventsEmit(a.ctx, "alerts:notify", map[string]interface{}{
		"title":   title,
		"message": message,
	})
}

// === Auto Insights Methods ===

// GetAutoInsights returns all auto-generated insights
func (a *App) GetAutoInsights(onlyUnread bool) []models.AutoInsight {
	if a.insightService == nil {
		return []models.AutoInsight{}
	}
	return a.insightService.GetInsights(onlyUnread)
}

// MarkInsightAsRead marks an insight as read
func (a *App) MarkInsightAsRead(insightID string) error {
	if a.insightService == nil {
		return fmt.Errorf("insight service not available")
	}
	return a.insightService.MarkAsRead(insightID)
}

// ClearOldInsights removes insights older than 24 hours
func (a *App) ClearOldInsights() error {
	if a.insightService == nil {
		return fmt.Errorf("insight service not available")
	}
	return a.insightService.ClearInsights()
}

// ClearAllInsights removes all insights
func (a *App) ClearAllInsights() error {
	if a.insightService == nil {
		return fmt.Errorf("insight service not available")
	}
	return a.insightService.ClearAllInsights()
}

// startInsightMonitoring starts background monitoring for auto insights
func (a *App) startInsightMonitoring() {
	a.mu.Lock()
	if a.insightMonitorOn || a.insightService == nil {
		a.mu.Unlock()
		return
	}
	a.insightMonitorOn = true
	a.mu.Unlock()

	// Check every 30 seconds for insights (less frequent to reduce noise)
	a.insightTicker = time.NewTicker(30 * time.Second)

	go func() {
		for range a.insightTicker.C {
			a.generateInsights()
		}
	}()
}

// stopInsightMonitoring stops the insight monitoring
func (a *App) stopInsightMonitoring() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.insightTicker != nil {
		a.insightTicker.Stop()
		a.insightMonitorOn = false
	}
}

// generateInsights collects system data and generates insights
func (a *App) generateInsights() {
	if a.insightService == nil {
		return
	}

	// Collect current system data
	stats, err := a.collector.GetDetailedStats()
	if err != nil {
		return
	}

	processes, err := a.collector.GetProcesses()
	if err != nil {
		processes = []models.ProcessInfo{}
	}

	security, err := a.collector.GetSecurityInfo()
	if err != nil {
		security = nil
	}

	// Generate insights based on current data
	newInsights := a.insightService.AnalyzeSystem(*stats, processes, security)

	// Emit events for new insights to update frontend
	if len(newInsights) > 0 {
		for _, insight := range newInsights {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "insights:new", map[string]interface{}{
					"insight": insight,
				})
			}
		}
	}
}
