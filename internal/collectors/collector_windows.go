//go:build windows

package collectors

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"sysmind/internal/exec"
	"sysmind/internal/models"
	"sysmind/internal/services"
)

type windowsCollector struct {
	lastNetStats          map[int32]*netStats
	lastCheck             time.Time
	lastDiskIO            *disk.IOCountersStat
	lastDiskIOTime        time.Time
	lastNetIO             *gnet.IOCountersStat
	lastNetIOTime         time.Time
	mu                    sync.Mutex
	dockerService         *services.DockerService
	devEnvironmentService *services.DevEnvironmentService
}

type netStats struct {
	bytesSent uint64
	bytesRecv uint64
	timestamp time.Time
}

func newPlatformCollector() Collector {
	dockerService := services.NewDockerService()
	devEnvironmentService := services.NewDevEnvironmentService(dockerService)

	return &windowsCollector{
		lastNetStats:          make(map[int32]*netStats),
		lastCheck:             time.Now(),
		dockerService:         dockerService,
		devEnvironmentService: devEnvironmentService,
	}
}

func (c *windowsCollector) GetProcesses() ([]models.ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var result []models.ProcessInfo
	for _, p := range procs {
		name, _ := p.Name()
		cmdline, _ := p.Cmdline()
		cpuPct, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		status, _ := p.Status()
		username, _ := p.Username()
		ppid, _ := p.Ppid()
		createTime, _ := p.CreateTime()
		numThreads, _ := p.NumThreads()

		var memMB float64
		if memInfo != nil {
			memMB = float64(memInfo.RSS) / 1024 / 1024
		}

		statusStr := "unknown"
		if len(status) > 0 {
			statusStr = status[0]
		}

		result = append(result, models.ProcessInfo{
			PID:         p.Pid,
			Name:        name,
			CommandLine: cmdline,
			CPUPercent:  cpuPct,
			MemoryMB:    memMB,
			Status:      statusStr,
			Username:    username,
			ParentPID:   ppid,
			CreateTime:  createTime,
			NumThreads:  numThreads,
		})
	}

	return result, nil
}

func (c *windowsCollector) GetPorts() ([]models.PortInfo, error) {
	var ports []models.PortInfo

	// TCP connections
	tcpPorts, _ := c.getNetstatPorts("TCP")
	ports = append(ports, tcpPorts...)

	// UDP connections
	udpPorts, _ := c.getNetstatPorts("UDP")
	ports = append(ports, udpPorts...)

	return ports, nil
}

func (c *windowsCollector) getNetstatPorts(protocol string) ([]models.PortInfo, error) {
	var ports []models.PortInfo

	var cmd *exec.Cmd
	if protocol == "TCP" {
		cmd = exec.Command("netstat", "-ano", "-p", "tcp")
	} else {
		cmd = exec.Command("netstat", "-ano", "-p", "udp")
	}

	output, err := cmd.Output()
	if err != nil {
		return ports, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Proto") || strings.HasPrefix(line, "Active") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		if !strings.HasPrefix(proto, protocol) {
			continue
		}

		localAddr := fields[1]
		remoteAddr := fields[2]

		var state string
		var pidStr string

		if protocol == "TCP" && len(fields) >= 5 {
			state = fields[3]
			pidStr = fields[4]
		} else if len(fields) >= 4 {
			state = "LISTENING"
			pidStr = fields[3]
		}

		pid, _ := strconv.ParseInt(pidStr, 10, 32)

		var port uint32
		if idx := strings.LastIndex(localAddr, ":"); idx != -1 {
			portStr := localAddr[idx+1:]
			if p, err := strconv.ParseUint(portStr, 10, 32); err == nil {
				port = uint32(p)
			}
		}

		// Get process name
		var processName string
		if pid > 0 {
			if p, err := process.NewProcess(int32(pid)); err == nil {
				processName, _ = p.Name()
			}
		}

		ports = append(ports, models.PortInfo{
			Port:        port,
			Protocol:    protocol,
			State:       state,
			LocalAddr:   localAddr,
			RemoteAddr:  remoteAddr,
			PID:         int32(pid),
			ProcessName: processName,
		})
	}

	return ports, nil
}

func (c *windowsCollector) GetNetworkUsage() ([]models.NetworkUsage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	elapsed := now.Sub(c.lastCheck).Seconds()
	if elapsed < 0.1 {
		elapsed = 0.1
	}

	var result []models.NetworkUsage
	newStats := make(map[int32]*netStats)

	for _, p := range procs {
		ioCounters, err := p.IOCounters()
		if err != nil {
			continue
		}

		name, _ := p.Name()
		bytesSent := ioCounters.WriteBytes
		bytesRecv := ioCounters.ReadBytes

		usage := models.NetworkUsage{
			PID:         p.Pid,
			ProcessName: name,
			BytesSent:   bytesSent,
			BytesRecv:   bytesRecv,
		}

		if last, ok := c.lastNetStats[p.Pid]; ok {
			sentDiff := bytesSent - last.bytesSent
			recvDiff := bytesRecv - last.bytesRecv
			usage.UploadSpeed = float64(sentDiff) / elapsed
			usage.DownloadSpeed = float64(recvDiff) / elapsed
		}

		newStats[p.Pid] = &netStats{
			bytesSent: bytesSent,
			bytesRecv: bytesRecv,
			timestamp: now,
		}

		if usage.BytesSent > 0 || usage.BytesRecv > 0 {
			result = append(result, usage)
		}
	}

	c.lastNetStats = newStats
	c.lastCheck = now

	return result, nil
}

func (c *windowsCollector) GetSystemStats() (float64, float64, error) {
	cpuPct, err := cpu.Percent(0, false)
	if err != nil {
		return 0, 0, err
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}

	var cpuUsage float64
	if len(cpuPct) > 0 {
		cpuUsage = cpuPct[0]
	}

	return cpuUsage, memStat.UsedPercent, nil
}

func (c *windowsCollector) GetDetailedStats() (*models.SystemStats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	stats := &models.SystemStats{
		Timestamp: now.UnixMilli(),
	}

	// CPU usage
	cpuPct, err := cpu.Percent(0, false)
	if err == nil && len(cpuPct) > 0 {
		stats.CPUPercent = cpuPct[0]
	}

	cpuPerCore, err := cpu.Percent(0, true)
	if err == nil {
		stats.CPUPerCore = cpuPerCore
	}

	// Memory stats
	memStat, err := mem.VirtualMemory()
	if err == nil {
		stats.MemoryPercent = memStat.UsedPercent
		stats.MemoryUsedGB = float64(memStat.Used) / 1024 / 1024 / 1024
		stats.MemoryTotalGB = float64(memStat.Total) / 1024 / 1024 / 1024
	}

	// Swap stats (Windows page file)
	swapStat, err := mem.SwapMemory()
	if err == nil {
		stats.SwapPercent = swapStat.UsedPercent
		stats.SwapUsedGB = float64(swapStat.Used) / 1024 / 1024 / 1024
		stats.SwapTotalGB = float64(swapStat.Total) / 1024 / 1024 / 1024
	}

	// Disk stats (C: drive)
	diskStat, err := disk.Usage("C:")
	if err == nil {
		stats.DiskPercent = diskStat.UsedPercent
		stats.DiskUsedGB = float64(diskStat.Used) / 1024 / 1024 / 1024
		stats.DiskTotalGB = float64(diskStat.Total) / 1024 / 1024 / 1024
	}

	// Disk I/O
	diskIO, err := disk.IOCounters()
	if err == nil {
		var totalRead, totalWrite uint64
		for _, io := range diskIO {
			totalRead += io.ReadBytes
			totalWrite += io.WriteBytes
		}
		if c.lastDiskIO != nil {
			elapsed := now.Sub(c.lastDiskIOTime).Seconds()
			if elapsed > 0 {
				stats.DiskReadSpeed = float64(totalRead-c.lastDiskIO.ReadBytes) / elapsed
				stats.DiskWriteSpeed = float64(totalWrite-c.lastDiskIO.WriteBytes) / elapsed
			}
		}
		c.lastDiskIO = &disk.IOCountersStat{ReadBytes: totalRead, WriteBytes: totalWrite}
		c.lastDiskIOTime = now
	}

	// Network I/O
	netIO, err := gnet.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		if c.lastNetIO != nil {
			elapsed := now.Sub(c.lastNetIOTime).Seconds()
			if elapsed > 0 {
				stats.NetUploadSpeed = float64(netIO[0].BytesSent-c.lastNetIO.BytesSent) / elapsed
				stats.NetDownSpeed = float64(netIO[0].BytesRecv-c.lastNetIO.BytesRecv) / elapsed
			}
		}
		c.lastNetIO = &netIO[0]
		c.lastNetIOTime = now
	}

	// Uptime
	uptime, err := host.Uptime()
	if err == nil {
		stats.Uptime = int64(uptime)
	}

	// Load average (not standard on Windows, use CPU as proxy)
	loadAvg, err := load.Avg()
	if err == nil {
		stats.LoadAvg1 = loadAvg.Load1
		stats.LoadAvg5 = loadAvg.Load5
		stats.LoadAvg15 = loadAvg.Load15
	}

	return stats, nil
}

func (c *windowsCollector) GetSecurityInfo() (*models.SecurityInfo, error) {
	info := &models.SecurityInfo{}

	// Check Windows Firewall status
	cmd := exec.Command("netsh", "advfirewall", "show", "allprofiles", "state")
	output, err := cmd.Output()
	if err == nil {
		outStr := string(output)
		if strings.Contains(outStr, "ON") {
			info.FirewallEnabled = true
			info.FirewallStatus = "Windows Firewall Active"
		} else {
			info.FirewallEnabled = false
			info.FirewallStatus = "Windows Firewall Inactive"
		}
	} else {
		info.FirewallStatus = "Unknown"
	}

	// Get port stats and external connections
	ports, err := c.GetPorts()
	if err == nil {
		info.OpenPorts = len(ports)

		// Build a map of PIDs to process names
		procNames := make(map[int32]string)
		procs, _ := c.GetProcesses()
		for _, p := range procs {
			procNames[p.PID] = p.Name
		}

		// Collect connection info
		for _, p := range ports {
			if p.State == "LISTENING" {
				info.ListeningPorts++
			}
			// Count external connections (not local, not 0.0.0.0)
			if p.RemoteAddr != "0.0.0.0:0" && !strings.HasPrefix(p.RemoteAddr, "127.") &&
				!strings.HasPrefix(p.RemoteAddr, "::1") {
				info.ExternalConns++

				// Add to unknown connections list for geo mapping
				conn := models.ConnectionInfo{
					LocalAddr:   p.LocalAddr,
					RemoteAddr:  p.RemoteAddr,
					ProcessName: p.ProcessName,
					PID:         p.PID,
				}

				// Extract IP from address
				if remoteIP := extractIPFromAddr(p.RemoteAddr); remoteIP != "" {
					conn.RemoteHost = remoteIP
				}

				info.UnknownConns = append(info.UnknownConns, conn)
			}
		}
	}

	// Detect suspicious processes
	info.SuspiciousProcs = c.detectSuspiciousProcesses()

	return info, nil
}

func (c *windowsCollector) detectSuspiciousProcesses() []models.SuspiciousProc {
	var suspicious []models.SuspiciousProc

	procs, err := process.Processes()
	if err != nil {
		return suspicious
	}

	suspiciousNames := []string{
		"cryptominer", "xmrig", "minerd", "cgminer", "bfgminer",
		"mimikatz", "psexec", "nc.exe", "ncat",
	}

	for _, p := range procs {
		name, _ := p.Name()
		cpuPct, _ := p.CPUPercent()

		var reasons []string
		riskLevel := "low"

		nameLower := strings.ToLower(name)
		for _, sus := range suspiciousNames {
			if strings.Contains(nameLower, sus) {
				reasons = append(reasons, fmt.Sprintf("Suspicious name pattern: %s", sus))
				riskLevel = "high"
			}
		}

		if cpuPct > 80 {
			reasons = append(reasons, fmt.Sprintf("High CPU usage: %.1f%%", cpuPct))
			if riskLevel != "high" {
				riskLevel = "medium"
			}
		}

		if len(reasons) > 0 {
			suspicious = append(suspicious, models.SuspiciousProc{
				PID:       p.Pid,
				Name:      name,
				Reasons:   reasons,
				RiskLevel: riskLevel,
			})
		}
	}

	return suspicious
}

func (c *windowsCollector) KillProcess(pid int32) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %d", pid)
	}
	return proc.Kill()
}

func (c *windowsCollector) SetProcessPriority(pid int32, priority int) error {
	// Windows uses different priority classes
	// Map Unix-style priority to Windows priority class
	var priorityClass string
	switch {
	case priority <= -10:
		priorityClass = "HIGH"
	case priority <= 0:
		priorityClass = "ABOVE_NORMAL"
	case priority <= 10:
		priorityClass = "BELOW_NORMAL"
	default:
		priorityClass = "IDLE"
	}

	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "CALL", "setpriority", priorityClass)
	return cmd.Run()
}

func (c *windowsCollector) GetDevEnvironmentInfo() (*models.DevEnvironmentInfo, error) {
	// Get current processes and ports
	processes, err := c.GetProcesses()
	if err != nil {
		// Return empty info instead of nil
		return &models.DevEnvironmentInfo{
			Containers:   []models.DockerContainer{},
			Environments: []models.DevEnvironment{},
			DevPorts:     []models.DevPort{},
		}, nil
	}

	ports, err := c.GetPorts()
	if err != nil {
		// Return empty info instead of nil
		return &models.DevEnvironmentInfo{
			Containers:   []models.DockerContainer{},
			Environments: []models.DevEnvironment{},
			DevPorts:     []models.DevPort{},
		}, nil
	}

	// Use dev environment service to analyze and detect environments
	info, err := c.devEnvironmentService.GetDevEnvironmentInfo(processes, ports)
	if err != nil || info == nil {
		return &models.DevEnvironmentInfo{
			Containers:   []models.DockerContainer{},
			Environments: []models.DevEnvironment{},
			DevPorts:     []models.DevPort{},
		}, nil
	}
	return info, nil
}
