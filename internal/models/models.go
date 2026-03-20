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

// IncidentProcessSample stores a lightweight process snapshot for rewind analysis
type IncidentProcessSample struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpuPercent"`
	MemoryMB   float64 `json:"memoryMB"`
	NumThreads int32   `json:"numThreads"`
	Status     string  `json:"status"`
}

// IncidentSample stores a synchronized system snapshot for rewind analysis
type IncidentSample struct {
	Timestamp      int64                   `json:"timestamp"`
	CPUPercent     float64                 `json:"cpuPercent"`
	MemoryPercent  float64                 `json:"memoryPercent"`
	DiskPercent    float64                 `json:"diskPercent"`
	DiskUsedGB     float64                 `json:"diskUsedGB"`
	DiskTotalGB    float64                 `json:"diskTotalGB"`
	NetUploadSpeed float64                 `json:"netUploadSpeed"`
	NetDownSpeed   float64                 `json:"netDownSpeed"`
	LoadAvg1       float64                 `json:"loadAvg1"`
	Processes      []IncidentProcessSample `json:"processes"`
}

// IncidentFinding represents a detected anomaly in the rewind window
type IncidentFinding struct {
	ID                  string  `json:"id"`
	Category            string  `json:"category"`
	Severity            string  `json:"severity"`
	Title               string  `json:"title"`
	Summary             string  `json:"summary"`
	Metric              string  `json:"metric"`
	StartedAt           int64   `json:"startedAt"`
	PeakAt              int64   `json:"peakAt"`
	StartValue          float64 `json:"startValue"`
	PeakValue           float64 `json:"peakValue"`
	CulpritPID          int32   `json:"culpritPid"`
	CulpritName         string  `json:"culpritName"`
	CulpritCPUPercent   float64 `json:"culpritCpuPercent"`
	CulpritMemoryMB     float64 `json:"culpritMemoryMB"`
	CulpritThreads      int32   `json:"culpritThreads"`
	CulpritStatus       string  `json:"culpritStatus"`
	ThreadHint          string  `json:"threadHint"`
	SyscallHint         string  `json:"syscallHint"`
	Confidence          float64 `json:"confidence"`
	ExactTraceAvailable bool    `json:"exactTraceAvailable"`
}

// IncidentRewind bundles rewind samples with detected findings
type IncidentRewind struct {
	WindowMinutes     int               `json:"windowMinutes"`
	ResolutionSeconds int               `json:"resolutionSeconds"`
	Samples           []IncidentSample  `json:"samples"`
	Findings          []IncidentFinding `json:"findings"`
	HighlightedAt     int64             `json:"highlightedAt"`
	Summary           string            `json:"summary"`
}

// TimeMachineSample stores persisted lower-frequency telemetry for past and future analysis.
type TimeMachineSample struct {
	Timestamp      int64                   `json:"timestamp"`
	CPUPercent     float64                 `json:"cpuPercent"`
	MemoryPercent  float64                 `json:"memoryPercent"`
	DiskPercent    float64                 `json:"diskPercent"`
	DiskUsedGB     float64                 `json:"diskUsedGB"`
	DiskTotalGB    float64                 `json:"diskTotalGB"`
	NetUploadSpeed float64                 `json:"netUploadSpeed"`
	NetDownSpeed   float64                 `json:"netDownSpeed"`
	LoadAvg1       float64                 `json:"loadAvg1"`
	Processes      []IncidentProcessSample `json:"processes"`
}

// TimeMachineAnnotation marks a notable event on the time machine timeline.
type TimeMachineAnnotation struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Severity    string  `json:"severity"`
	Timestamp   int64   `json:"timestamp"`
	Title       string  `json:"title"`
	Summary     string  `json:"summary"`
	ProcessName string  `json:"processName"`
	ProcessPID  int32   `json:"processPid"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
}

// TimeMachineForecast represents a forward projection based on recent telemetry.
type TimeMachineForecast struct {
	ID             string  `json:"id"`
	Kind           string  `json:"kind"`
	Severity       string  `json:"severity"`
	Title          string  `json:"title"`
	Summary        string  `json:"summary"`
	PredictedAt    int64   `json:"predictedAt"`
	CurrentValue   float64 `json:"currentValue"`
	ProjectedValue float64 `json:"projectedValue"`
	Confidence     float64 `json:"confidence"`
	Unit           string  `json:"unit"`
}

// TimeMachineView bundles persisted history, annotations, and forecasts for the scrubber UI.
type TimeMachineView struct {
	WindowHours        int                     `json:"windowHours"`
	RetentionHours     int                     `json:"retentionHours"`
	SamplingSeconds    int                     `json:"samplingSeconds"`
	Samples            []TimeMachineSample     `json:"samples"`
	Annotations        []TimeMachineAnnotation `json:"annotations"`
	Forecasts          []TimeMachineForecast   `json:"forecasts"`
	Summary            string                  `json:"summary"`
	LastUpdated        int64                   `json:"lastUpdated"`
	PersistenceEnabled bool                    `json:"persistenceEnabled"`
}

// BaselineDriftFinding describes behavior that deviates from this machine's learned baseline.
type BaselineDriftFinding struct {
	ID            string  `json:"id"`
	Kind          string  `json:"kind"`
	Category      string  `json:"category"`
	Severity      string  `json:"severity"`
	Title         string  `json:"title"`
	Summary       string  `json:"summary"`
	Metric        string  `json:"metric"`
	Unit          string  `json:"unit"`
	CurrentValue  float64 `json:"currentValue"`
	BaselineValue float64 `json:"baselineValue"`
	ExpectedHigh  float64 `json:"expectedHigh"`
	DeltaPercent  float64 `json:"deltaPercent"`
	ProcessName   string  `json:"processName,omitempty"`
	ProcessPID    int32   `json:"processPid,omitempty"`
	Port          uint32  `json:"port,omitempty"`
	Protocol      string  `json:"protocol,omitempty"`
	FirstSeenAt   int64   `json:"firstSeenAt,omitempty"`
	LastSeenAt    int64   `json:"lastSeenAt,omitempty"`
	SampleCount   int     `json:"sampleCount"`
	Confidence    float64 `json:"confidence"`
	IsNew         bool    `json:"isNew"`
}

// BaselineDriftView bundles the current new/unusual findings against learned behavior.
type BaselineDriftView struct {
	GeneratedAt   int64                  `json:"generatedAt"`
	CoverageHours int                    `json:"coverageHours"`
	SampleCount   int                    `json:"sampleCount"`
	Learning      bool                   `json:"learning"`
	Findings      []BaselineDriftFinding `json:"findings"`
	Summary       string                 `json:"summary"`
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

// PrivacyConfig contains privacy settings for AI data sharing
type PrivacyConfig struct {
	// What data to share with AI providers
	ShareProcessNames    bool `json:"shareProcessNames"`    // Share process names (e.g., "firefox", "code")
	ShareProcessDetails  bool `json:"shareProcessDetails"`  // Share CPU/memory usage per process
	ShareNetworkPorts    bool `json:"shareNetworkPorts"`    // Share open ports and listening services
	ShareConnectionIPs   bool `json:"shareConnectionIPs"`   // Share remote IP addresses
	ShareConnectionGeo   bool `json:"shareConnectionGeo"`   // Share geographic location of connections
	ShareSecurityInfo    bool `json:"shareSecurityInfo"`    // Share security alerts and suspicious process info
	ShareSystemStats     bool `json:"shareSystemStats"`     // Share CPU/Memory/Disk percentages (basic stats)
	AnonymizeProcesses   bool `json:"anonymizeProcesses"`   // Replace process names with categories
	AnonymizeConnections bool `json:"anonymizeConnections"` // Replace IPs with provider categories (e.g., "CDN", "Cloud")
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
