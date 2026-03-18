package collectors

import (
	"strings"
	"sysmind/internal/models"
)

// Collector interface defines methods for system data collection
type Collector interface {
	GetProcesses() ([]models.ProcessInfo, error)
	GetPorts() ([]models.PortInfo, error)
	GetNetworkUsage() ([]models.NetworkUsage, error)
	GetSystemStats() (cpuUsage float64, memUsage float64, err error)
	GetDetailedStats() (*models.SystemStats, error)
	GetSecurityInfo() (*models.SecurityInfo, error)
	GetDevEnvironmentInfo() (*models.DevEnvironmentInfo, error)
	KillProcess(pid int32) error
	SetProcessPriority(pid int32, priority int) error
}

// NewCollector creates a platform-specific collector
func NewCollector() Collector {
	return newPlatformCollector()
}

// extractIPFromAddr extracts IP from "ip:port" format
func extractIPFromAddr(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		return addr[:idx]
	}
	return addr
}
