package models

import "time"

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	CommandLine string  `json:"commandLine"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemoryMB    float64 `json:"memoryMB"`
	Status      string  `json:"status"`
	Username    string  `json:"username"`
	ParentPID   int32   `json:"parentPid"`
	CreateTime  int64   `json:"createTime"`
	NumThreads  int32   `json:"numThreads"`
}

// PortInfo contains information about an open port
type PortInfo struct {
	Port        uint32 `json:"port"`
	Protocol    string `json:"protocol"`
	State       string `json:"state"`
	LocalAddr   string `json:"localAddr"`
	RemoteAddr  string `json:"remoteAddr"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"processName"`
}

// NetworkUsage contains bandwidth usage for a process
type NetworkUsage struct {
	PID           int32   `json:"pid"`
	ProcessName   string  `json:"processName"`
	BytesSent     uint64  `json:"bytesSent"`
	BytesRecv     uint64  `json:"bytesRecv"`
	UploadSpeed   float64 `json:"uploadSpeed"`   // bytes per second
	DownloadSpeed float64 `json:"downloadSpeed"` // bytes per second
}

// SystemStats contains detailed system statistics
type SystemStats struct {
	CPUPercent     float64   `json:"cpuPercent"`
	CPUPerCore     []float64 `json:"cpuPerCore"`
	MemoryPercent  float64   `json:"memoryPercent"`
	MemoryUsedGB   float64   `json:"memoryUsedGB"`
	MemoryTotalGB  float64   `json:"memoryTotalGB"`
	SwapPercent    float64   `json:"swapPercent"`
	SwapUsedGB     float64   `json:"swapUsedGB"`
	SwapTotalGB    float64   `json:"swapTotalGB"`
	DiskPercent    float64   `json:"diskPercent"`
	DiskUsedGB     float64   `json:"diskUsedGB"`
	DiskTotalGB    float64   `json:"diskTotalGB"`
	DiskReadSpeed  float64   `json:"diskReadSpeed"`  // bytes per second
	DiskWriteSpeed float64   `json:"diskWriteSpeed"` // bytes per second
	NetUploadSpeed float64   `json:"netUploadSpeed"` // bytes per second
	NetDownSpeed   float64   `json:"netDownSpeed"`   // bytes per second
	Uptime         int64     `json:"uptime"`         // seconds
	LoadAvg1       float64   `json:"loadAvg1"`
	LoadAvg5       float64   `json:"loadAvg5"`
	LoadAvg15      float64   `json:"loadAvg15"`
	Timestamp      int64     `json:"timestamp"`
}

// ResourceTimelinePoint stores a historical system snapshot for trend analysis
type ResourceTimelinePoint struct {
	Timestamp      int64   `json:"timestamp"`
	CPUPercent     float64 `json:"cpuPercent"`
	MemoryPercent  float64 `json:"memoryPercent"`
	DiskPercent    float64 `json:"diskPercent"`
	NetUploadSpeed float64 `json:"netUploadSpeed"`
	NetDownSpeed   float64 `json:"netDownSpeed"`
}

// SystemContext contains all system monitoring data for AI analysis
type SystemContext struct {
	Processes    []ProcessInfo  `json:"processes"`
	Ports        []PortInfo     `json:"ports"`
	Network      []NetworkUsage `json:"network"`
	Timestamp    time.Time      `json:"timestamp"`
	CPUUsage     float64        `json:"cpuUsage"`
	MemUsage     float64        `json:"memUsage"`
	DiskUsage    float64        `json:"diskUsage"`
	DiskUsedGB   float64        `json:"diskUsedGB"`
	DiskTotalGB  float64        `json:"diskTotalGB"`
	SecurityInfo *SecurityInfo  `json:"securityInfo,omitempty"`
}

// Alert represents a system alert
type Alert struct {
	ID        string `json:"id"`
	Type      string `json:"type"`     // "cpu", "memory", "disk", "process", "network", "security"
	Severity  string `json:"severity"` // "info", "warning", "critical"
	Title     string `json:"title"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Dismissed bool   `json:"dismissed"`
	Data      any    `json:"data,omitempty"`
}

// AlertConfig contains alert threshold settings
type AlertConfig struct {
	CPUThreshold      float64 `json:"cpuThreshold"`
	MemoryThreshold   float64 `json:"memoryThreshold"`
	DiskThreshold     float64 `json:"diskThreshold"`
	EnableAlerts      bool    `json:"enableAlerts"`
	EnableSound       bool    `json:"enableSound"`
	EnableDesktopNotf bool    `json:"enableDesktopNotf"`
}

// SecurityInfo contains security-related information
type SecurityInfo struct {
	FirewallEnabled bool             `json:"firewallEnabled"`
	FirewallStatus  string           `json:"firewallStatus"`
	SuspiciousProcs []SuspiciousProc `json:"suspiciousProcs"`
	OpenPorts       int              `json:"openPorts"`
	ListeningPorts  int              `json:"listeningPorts"`
	ExternalConns   int              `json:"externalConns"`
	UnknownConns    []ConnectionInfo `json:"unknownConns"`
}

// SuspiciousProc represents a potentially suspicious process
type SuspiciousProc struct {
	PID       int32    `json:"pid"`
	Name      string   `json:"name"`
	Reasons   []string `json:"reasons"`
	RiskLevel string   `json:"riskLevel"` // "low", "medium", "high"
}

// ConnectionInfo contains connection details
type ConnectionInfo struct {
	LocalAddr   string  `json:"localAddr"`
	RemoteAddr  string  `json:"remoteAddr"`
	RemoteHost  string  `json:"remoteHost"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ProcessName string  `json:"processName"`
	PID         int32   `json:"pid"`
}

// PromptTemplate represents a saved prompt template
type PromptTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
}

// ChatMessage represents a message in the AI chat
type ChatMessage struct {
	ID        string `json:"id"`
	Role      string `json:"role"` // "user", "assistant", or "error"
	Content   string `json:"content"`
	RiskLevel string `json:"riskLevel,omitempty"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp in milliseconds
}

// ChatSession represents a chat session with history
type ChatSession struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt int64         `json:"createdAt"` // Unix timestamp in milliseconds
	UpdatedAt int64         `json:"updatedAt"` // Unix timestamp in milliseconds
}

// ChatSessionSummary is a lightweight version for listing sessions
type ChatSessionSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	MessageCount int    `json:"messageCount"`
}

// AIConfig contains AI provider configuration
type AIConfig struct {
	Provider       string `json:"provider"`       // "openai", "cloudflare", "local"
	Model          string `json:"model"`          // Model identifier
	APIKey         string `json:"apiKey"`         // API key for the provider
	CloudflareAcct string `json:"cloudflareAcct"` // Cloudflare account ID
	LocalEndpoint  string `json:"localEndpoint"`  // Local LLM endpoint (e.g., http://localhost:11434)
}

// AIResponse represents a structured response from the AI
type AIResponse struct {
	Explanation string   `json:"explanation"`
	RiskLevel   string   `json:"riskLevel"` // "low", "medium", "high"
	Suggestions []string `json:"suggestions"`
}

// AutoInsight represents an automatically generated system insight
type AutoInsight struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`       // e.g., "High CPU Usage Detected"
	Message     string   `json:"message"`     // e.g., "CPU has been above 80% for 10 minutes"
	Category    string   `json:"category"`    // "performance", "security", "network", "process"
	Severity    string   `json:"severity"`    // "info", "warning", "critical"
	Timestamp   int64    `json:"timestamp"`   // Unix timestamp in milliseconds
	IsRead      bool     `json:"isRead"`      // Whether user has seen this insight
	Data        string   `json:"data"`        // JSON string with relevant data
	ActionItems []string `json:"actionItems"` // Suggested actions
}

// InsightTrigger tracks conditions for generating insights
type InsightTrigger struct {
	Type       string `json:"type"`       // "cpu_high", "new_connections", "suspicious_process"
	StartTime  int64  `json:"startTime"`  // When condition started
	LastUpdate int64  `json:"lastUpdate"` // Last time condition was updated
	Count      int    `json:"count"`      // How many times triggered
	Data       string `json:"data"`       // Associated data
	Triggered  bool   `json:"triggered"`  // Whether insight was already generated
}

// DockerContainer represents a Docker container
type DockerContainer struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Status     string            `json:"status"` // running, exited, paused
	State      string            `json:"state"`  // created, restarting, running, removing, paused, exited, dead
	Ports      []ContainerPort   `json:"ports"`
	Labels     map[string]string `json:"labels"`
	Command    string            `json:"command"`
	CreatedAt  int64             `json:"createdAt"`
	StartedAt  int64             `json:"startedAt"`
	FinishedAt int64             `json:"finishedAt,omitempty"`
	ExitCode   int               `json:"exitCode,omitempty"`
	CPUPercent float64           `json:"cpuPercent"`
	MemoryMB   float64           `json:"memoryMB"`
	NetworkRX  uint64            `json:"networkRX"`
	NetworkTX  uint64            `json:"networkTX"`
}

// ContainerPort represents a port mapping for a container
type ContainerPort struct {
	PrivatePort uint16 `json:"privatePort"`
	PublicPort  uint16 `json:"publicPort,omitempty"`
	Type        string `json:"type"` // tcp, udp
	IP          string `json:"ip"`   // host IP
}

// DevEnvironment represents a detected development environment
type DevEnvironment struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`       // "Next.js App", "PostgreSQL Database"
	Type        string   `json:"type"`       // "web", "database", "api", "build", "proxy"
	Technology  string   `json:"technology"` // "nextjs", "postgres", "redis", "nginx", "webpack"
	Port        uint16   `json:"port"`
	ProcessName string   `json:"processName"`
	ProcessPID  int32    `json:"processPID"`
	ContainerID string   `json:"containerID,omitempty"`
	Status      string   `json:"status"`         // "running", "stopped", "building"
	Icon        string   `json:"icon"`           // emoji or icon identifier
	Description string   `json:"description"`    // detailed description
	URLs        []string `json:"urls,omitempty"` // accessible URLs
}

// DevEnvironmentInfo contains all development environment data
type DevEnvironmentInfo struct {
	Containers    []DockerContainer `json:"containers"`
	Environments  []DevEnvironment  `json:"environments"`
	DevPorts      []DevPort         `json:"devPorts"`
	DockerRunning bool              `json:"dockerRunning"`
}

// DevPort represents a development-related port with intelligent identification
type DevPort struct {
	Port        uint16 `json:"port"`
	ProcessName string `json:"processName"`
	ProcessPID  int32  `json:"processPID"`
	Technology  string `json:"technology"` // "nextjs", "react", "vue", "postgres", etc.
	Framework   string `json:"framework"`  // "Next.js", "Create React App", "Vue CLI", etc.
	Icon        string `json:"icon"`       // emoji
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
}
