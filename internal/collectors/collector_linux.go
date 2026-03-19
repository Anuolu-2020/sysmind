//go:build linux

package collectors

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

type linuxCollector struct {
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

	return &linuxCollector{
		lastNetStats:          make(map[int32]*netStats),
		lastCheck:             time.Now(),
		dockerService:         dockerService,
		devEnvironmentService: devEnvironmentService,
	}
}

func (c *linuxCollector) GetProcesses() ([]models.ProcessInfo, error) {
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

func (c *linuxCollector) GetPorts() ([]models.PortInfo, error) {
	var ports []models.PortInfo

	// Parse TCP connections
	tcpPorts, err := c.parseProcNet("/proc/net/tcp", "tcp")
	if err == nil {
		ports = append(ports, tcpPorts...)
	}

	// Parse TCP6 connections
	tcp6Ports, err := c.parseProcNet("/proc/net/tcp6", "tcp")
	if err == nil {
		ports = append(ports, tcp6Ports...)
	}

	// Parse UDP connections
	udpPorts, err := c.parseProcNet("/proc/net/udp", "udp")
	if err == nil {
		ports = append(ports, udpPorts...)
	}

	// Parse UDP6 connections
	udp6Ports, err := c.parseProcNet("/proc/net/udp6", "udp")
	if err == nil {
		ports = append(ports, udp6Ports...)
	}

	// Build inode to PID map
	inodeToPID := c.buildInodeToPIDMap()

	// Resolve process names
	for i := range ports {
		if pid, ok := inodeToPID[ports[i].PID]; ok {
			ports[i].PID = pid
			if p, err := process.NewProcess(pid); err == nil {
				name, _ := p.Name()
				ports[i].ProcessName = name
			}
		}
	}

	return ports, nil
}

func (c *linuxCollector) parseProcNet(path, protocol string) ([]models.PortInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []models.PortInfo
	scanner := bufio.NewScanner(file)
	scanner.Scan() // Skip header

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		localAddr := fields[1]
		remoteAddr := fields[2]
		state := fields[3]
		inode, _ := strconv.ParseInt(fields[9], 10, 32)

		localIP, localPort := parseHexAddr(localAddr)
		remoteIP, remotePort := parseHexAddr(remoteAddr)

		stateStr := tcpStateToString(state)
		if protocol == "udp" {
			if state == "07" {
				stateStr = "LISTENING"
			} else {
				stateStr = "ESTABLISHED"
			}
		}

		ports = append(ports, models.PortInfo{
			Port:       localPort,
			Protocol:   strings.ToUpper(protocol),
			State:      stateStr,
			LocalAddr:  fmt.Sprintf("%s:%d", localIP, localPort),
			RemoteAddr: fmt.Sprintf("%s:%d", remoteIP, remotePort),
			PID:        int32(inode), // Will be resolved later
		})
	}

	return ports, nil
}

func parseHexAddr(addr string) (string, uint32) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", 0
	}

	ipHex := parts[0]
	portHex := parts[1]

	port, _ := strconv.ParseUint(portHex, 16, 32)

	var ip net.IP
	if len(ipHex) == 8 {
		// IPv4
		ipBytes, _ := hex.DecodeString(ipHex)
		ip = net.IPv4(ipBytes[3], ipBytes[2], ipBytes[1], ipBytes[0])
	} else {
		// IPv6
		ipBytes, _ := hex.DecodeString(ipHex)
		ip = make(net.IP, 16)
		for i := 0; i < 16; i += 4 {
			ip[i], ip[i+1], ip[i+2], ip[i+3] = ipBytes[i+3], ipBytes[i+2], ipBytes[i+1], ipBytes[i]
		}
	}

	return ip.String(), uint32(port)
}

func tcpStateToString(state string) string {
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTENING",
		"0B": "CLOSING",
	}
	if s, ok := states[strings.ToUpper(state)]; ok {
		return s
	}
	return "UNKNOWN"
}

func (c *linuxCollector) buildInodeToPIDMap() map[int32]int32 {
	result := make(map[int32]int32)

	procDirs, _ := filepath.Glob("/proc/[0-9]*/fd/*")
	for _, fdPath := range procDirs {
		link, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}

		if strings.HasPrefix(link, "socket:[") {
			inodeStr := strings.TrimPrefix(link, "socket:[")
			inodeStr = strings.TrimSuffix(inodeStr, "]")
			inode, _ := strconv.ParseInt(inodeStr, 10, 32)

			// Extract PID from path
			parts := strings.Split(fdPath, "/")
			if len(parts) >= 3 {
				pid, _ := strconv.ParseInt(parts[2], 10, 32)
				result[int32(inode)] = int32(pid)
			}
		}
	}

	return result
}

func (c *linuxCollector) GetNetworkUsage() ([]models.NetworkUsage, error) {
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

		// Only include processes with network activity
		if usage.BytesSent > 0 || usage.BytesRecv > 0 {
			result = append(result, usage)
		}
	}

	c.lastNetStats = newStats
	c.lastCheck = now

	return result, nil
}

func (c *linuxCollector) GetSystemStats() (float64, float64, error) {
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

func (c *linuxCollector) GetDetailedStats() (*models.SystemStats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	stats := &models.SystemStats{
		Timestamp: now.UnixMilli(),
	}

	// CPU usage (overall and per-core)
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

	// Swap stats
	swapStat, err := mem.SwapMemory()
	if err == nil {
		stats.SwapPercent = swapStat.UsedPercent
		stats.SwapUsedGB = float64(swapStat.Used) / 1024 / 1024 / 1024
		stats.SwapTotalGB = float64(swapStat.Total) / 1024 / 1024 / 1024
	}

	// Disk stats
	diskStat, err := disk.Usage("/")
	if err == nil {
		stats.DiskPercent = diskStat.UsedPercent
		stats.DiskUsedGB = float64(diskStat.Used) / 1024 / 1024 / 1024
		stats.DiskTotalGB = float64(diskStat.Total) / 1024 / 1024 / 1024
	}

	// Disk I/O speed
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

	// Network I/O speed
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

	// Load average
	loadAvg, err := load.Avg()
	if err == nil {
		stats.LoadAvg1 = loadAvg.Load1
		stats.LoadAvg5 = loadAvg.Load5
		stats.LoadAvg15 = loadAvg.Load15
	}

	return stats, nil
}

func (c *linuxCollector) GetSecurityInfo() (*models.SecurityInfo, error) {
	info := &models.SecurityInfo{}

	// Check firewall status (ufw or iptables)
	info.FirewallEnabled, info.FirewallStatus = c.checkFirewallStatus()

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
			if p.RemoteAddr != "0.0.0.0:0" && p.RemoteAddr != "[::]:0" &&
				!strings.HasPrefix(p.RemoteAddr, "127.") && !strings.HasPrefix(p.RemoteAddr, "::1") {
				info.ExternalConns++

				// Add to unknown connections list for geo mapping
				conn := models.ConnectionInfo{
					LocalAddr:   p.LocalAddr,
					RemoteAddr:  p.RemoteAddr,
					ProcessName: p.ProcessName,
					PID:         p.PID,
				}

				// Try to resolve hostname
				if remoteIP := extractIPFromAddr(p.RemoteAddr); remoteIP != "" {
					// Optional: could do reverse DNS lookup here, but it's slow
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

func (c *linuxCollector) checkFirewallStatus() (bool, string) {
	// Try ufw first
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err == nil {
		outStr := string(output)
		if strings.Contains(outStr, "Status: active") {
			return true, "UFW Active"
		} else if strings.Contains(outStr, "Status: inactive") {
			return false, "UFW Inactive"
		}
	}

	// Try iptables
	cmd = exec.Command("iptables", "-L", "-n")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		ruleCount := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "ACCEPT") || strings.HasPrefix(line, "DROP") || strings.HasPrefix(line, "REJECT") {
				ruleCount++
			}
		}
		if ruleCount > 0 {
			return true, fmt.Sprintf("iptables (%d rules)", ruleCount)
		}
		return false, "iptables (no rules)"
	}

	return false, "Unknown"
}

func (c *linuxCollector) detectSuspiciousProcesses() []models.SuspiciousProc {
	var suspicious []models.SuspiciousProc

	procs, err := process.Processes()
	if err != nil {
		return suspicious
	}

	// Known suspicious patterns
	suspiciousNames := []string{
		"cryptominer", "xmrig", "minerd", "cgminer", "bfgminer",
		"kworker", "kdevtmpfs", // when not actual kernel threads
	}

	suspiciousPorts := map[uint32]string{
		4444:  "Common mining pool port",
		3333:  "Common mining pool port",
		6666:  "Potential backdoor",
		31337: "Elite backdoor",
	}

	ports, _ := c.GetPorts()
	portByPID := make(map[int32][]uint32)
	for _, p := range ports {
		portByPID[p.PID] = append(portByPID[p.PID], p.Port)
	}

	for _, p := range procs {
		name, _ := p.Name()
		cmdline, _ := p.Cmdline()
		cpuPct, _ := p.CPUPercent()
		username, _ := p.Username()

		var reasons []string
		riskLevel := "low"

		// Check for suspicious names
		nameLower := strings.ToLower(name)
		for _, sus := range suspiciousNames {
			if strings.Contains(nameLower, sus) {
				reasons = append(reasons, fmt.Sprintf("Suspicious name pattern: %s", sus))
				riskLevel = "high"
			}
		}

		// High CPU usage by unknown process
		if cpuPct > 80 && username != "root" && !isKnownProcess(name) {
			reasons = append(reasons, fmt.Sprintf("High CPU usage: %.1f%%", cpuPct))
			if riskLevel != "high" {
				riskLevel = "medium"
			}
		}

		// Process running on suspicious port
		if procPorts, ok := portByPID[p.Pid]; ok {
			for _, port := range procPorts {
				if desc, isSus := suspiciousPorts[port]; isSus {
					reasons = append(reasons, fmt.Sprintf("Suspicious port %d: %s", port, desc))
					riskLevel = "high"
				}
			}
		}

		// Hidden command line (common for malware)
		if name != "" && cmdline == "" {
			// Check if it's a kernel thread (those legitimately have no cmdline)
			if !strings.HasPrefix(name, "[") {
				reasons = append(reasons, "Hidden command line")
				if riskLevel == "low" {
					riskLevel = "medium"
				}
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

func isKnownProcess(name string) bool {
	known := map[string]bool{
		"chrome": true, "firefox": true, "code": true, "node": true,
		"go": true, "python": true, "java": true, "npm": true,
		"systemd": true, "dbus-daemon": true, "pulseaudio": true,
		"Xorg": true, "gnome-shell": true, "kwin": true,
		"wails": true, "sysmind": true,
	}
	return known[name]
}

func (c *linuxCollector) KillProcess(pid int32) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %d", pid)
	}
	return proc.Kill()
}

func (c *linuxCollector) SetProcessPriority(pid int32, priority int) error {
	// Priority should be between -20 (highest) and 19 (lowest)
	if priority < -20 || priority > 19 {
		return fmt.Errorf("priority must be between -20 and 19")
	}
	return syscall.Setpriority(syscall.PRIO_PROCESS, int(pid), priority)
}

func (c *linuxCollector) GetDevEnvironmentInfo() (*models.DevEnvironmentInfo, error) {
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
