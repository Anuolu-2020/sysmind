package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"sysmind/internal/models"
)

// DockerService provides Docker container information with caching
type DockerService struct {
	cache       map[string]interface{}
	cacheMutex  sync.RWMutex
	cacheExpiry map[string]time.Time
	cacheTTL    time.Duration
}

// NewDockerService creates a new Docker service with caching
func NewDockerService() *DockerService {
	return &DockerService{
		cache:       make(map[string]interface{}),
		cacheExpiry: make(map[string]time.Time),
		cacheTTL:    5 * time.Second, // Cache for 5 seconds
	}
}

// getCached retrieves a cached value if it's still valid
func (ds *DockerService) getCached(key string) (interface{}, bool) {
	ds.cacheMutex.RLock()
	defer ds.cacheMutex.RUnlock()

	value, exists := ds.cache[key]
	if !exists {
		return nil, false
	}

	expiry, hasExpiry := ds.cacheExpiry[key]
	if hasExpiry && time.Now().After(expiry) {
		// Cache expired, remove it
		delete(ds.cache, key)
		delete(ds.cacheExpiry, key)
		return nil, false
	}

	return value, true
}

// setCached stores a value in cache with expiry
func (ds *DockerService) setCached(key string, value interface{}) {
	ds.cacheMutex.Lock()
	defer ds.cacheMutex.Unlock()

	ds.cache[key] = value
	ds.cacheExpiry[key] = time.Now().Add(ds.cacheTTL)
}

// clearCache clears all cached data
func (ds *DockerService) clearCache() {
	ds.cacheMutex.Lock()
	defer ds.cacheMutex.Unlock()

	ds.cache = make(map[string]interface{})
	ds.cacheExpiry = make(map[string]time.Time)
}

// IsDockerAvailable checks if Docker is installed and running (cached)
func (ds *DockerService) IsDockerAvailable() bool {
	cacheKey := "docker_available"
	if cached, found := ds.getCached(cacheKey); found {
		return cached.(bool)
	}

	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	available := cmd.Run() == nil

	ds.setCached(cacheKey, available)
	return available
}

// GetContainers retrieves all Docker containers (cached)
func (ds *DockerService) GetContainers() ([]models.DockerContainer, error) {
	cacheKey := "containers_list"
	if cached, found := ds.getCached(cacheKey); found {
		return cached.([]models.DockerContainer), nil
	}

	if !ds.IsDockerAvailable() {
		return []models.DockerContainer{}, nil
	}

	// Get container list and stats in one operation
	containers, err := ds.getContainerListWithStats()
	if err != nil {
		return nil, err
	}

	ds.setCached(cacheKey, containers)
	return containers, nil
}

// getContainerListWithStats gets containers and their stats efficiently
func (ds *DockerService) getContainerListWithStats() ([]models.DockerContainer, error) {
	// Get container list
	cmd := exec.Command("docker", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.Command}}\t{{.CreatedAt}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get containers: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return []models.DockerContainer{}, nil // No containers or only header
	}

	var containers []models.DockerContainer
	var runningContainers []string

	// Parse container list first
	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		container := ds.parseContainerLine(line)
		if container.ID != "" {
			// Enrich with detailed info
			ds.enrichContainerInfo(&container)
			containers = append(containers, container)

			if container.State == "running" {
				runningContainers = append(runningContainers, container.ID)
			}
		}
	}

	// Get stats for all running containers in batch
	if len(runningContainers) > 0 {
		ds.batchGetContainerStats(containers, runningContainers)
	}

	return containers, nil
}

// batchGetContainerStats gets stats for multiple containers efficiently
func (ds *DockerService) batchGetContainerStats(containers []models.DockerContainer, runningIDs []string) {
	if len(runningIDs) == 0 {
		return
	}

	// Get stats for all running containers in one command
	args := append([]string{"stats", "--no-stream", "--format", "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"}, runningIDs...)
	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return
	}

	// Create a map for quick lookup
	containerMap := make(map[string]*models.DockerContainer)
	for i := range containers {
		if containers[i].State == "running" {
			containerMap[containers[i].ID] = &containers[i]
		}
	}

	// Parse stats and update containers
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			containerID := strings.TrimSpace(parts[0])

			// Find container (try both full ID and short ID)
			var container *models.DockerContainer
			if c, exists := containerMap[containerID]; exists {
				container = c
			} else {
				// Try short ID match
				for id, c := range containerMap {
					if strings.HasPrefix(id, containerID) {
						container = c
						break
					}
				}
			}

			if container != nil {
				// Parse CPU percentage
				cpuStr := strings.TrimSpace(strings.Replace(parts[1], "%", "", -1))
				if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
					container.CPUPercent = cpu
				}

				// Parse memory usage
				memStr := strings.TrimSpace(parts[2])
				container.MemoryMB = ds.parseMemoryUsage(memStr)

				// Parse network I/O
				netStr := strings.TrimSpace(parts[3])
				rx, tx := ds.parseNetworkIO(netStr)
				container.NetworkRX = rx
				container.NetworkTX = tx
			}
		}
	}
}

// parseContainerLine parses a single container line from docker ps output
func (ds *DockerService) parseContainerLine(line string) models.DockerContainer {
	parts := strings.Split(line, "\t")
	if len(parts) < 6 {
		return models.DockerContainer{}
	}

	container := models.DockerContainer{
		ID:      strings.TrimSpace(parts[0]),
		Name:    strings.TrimSpace(parts[1]),
		Image:   strings.TrimSpace(parts[2]),
		Status:  strings.TrimSpace(parts[3]),
		Command: strings.TrimSpace(parts[5]),
	}

	// Parse ports
	portStr := strings.TrimSpace(parts[4])
	container.Ports = ds.parsePorts(portStr)

	// Set state based on status
	if strings.Contains(strings.ToLower(container.Status), "up") {
		container.State = "running"
	} else if strings.Contains(strings.ToLower(container.Status), "exited") {
		container.State = "exited"
	} else {
		container.State = "unknown"
	}

	return container
}

// enrichContainerInfo adds detailed information about a container
func (ds *DockerService) enrichContainerInfo(container *models.DockerContainer) {
	// Get detailed container info
	cmd := exec.Command("docker", "inspect", container.ID)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil || len(inspectData) == 0 {
		return
	}

	data := inspectData[0]

	// Extract timestamps
	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			container.CreatedAt = t.Unix()
		}
	}

	if state, ok := data["State"].(map[string]interface{}); ok {
		if started, ok := state["StartedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, started); err == nil {
				container.StartedAt = t.Unix()
			}
		}
		if finished, ok := state["FinishedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, finished); err == nil {
				container.FinishedAt = t.Unix()
			}
		}
		if exitCode, ok := state["ExitCode"].(float64); ok {
			container.ExitCode = int(exitCode)
		}
	}

	// Extract labels
	if config, ok := data["Config"].(map[string]interface{}); ok {
		if labels, ok := config["Labels"].(map[string]interface{}); ok {
			container.Labels = make(map[string]string)
			for k, v := range labels {
				if str, ok := v.(string); ok {
					container.Labels[k] = str
				}
			}
		}
	}

	// Get container stats if running
	if container.State == "running" {
		ds.getContainerStats(container)
	}
}

// getContainerStats retrieves CPU and memory stats for a running container
func (ds *DockerService) getContainerStats(container *models.DockerContainer) {
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "table {{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}", container.ID)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return
	}

	parts := strings.Split(lines[1], "\t")
	if len(parts) >= 3 {
		// Parse CPU percentage
		cpuStr := strings.TrimSpace(strings.Replace(parts[0], "%", "", -1))
		if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
			container.CPUPercent = cpu
		}

		// Parse memory usage
		memStr := strings.TrimSpace(parts[1])
		container.MemoryMB = ds.parseMemoryUsage(memStr)

		// Parse network I/O
		netStr := strings.TrimSpace(parts[2])
		rx, tx := ds.parseNetworkIO(netStr)
		container.NetworkRX = rx
		container.NetworkTX = tx
	}
}

// parsePorts parses port mappings from docker ps output
func (ds *DockerService) parsePorts(portStr string) []models.ContainerPort {
	if portStr == "" {
		return []models.ContainerPort{}
	}

	var ports []models.ContainerPort
	portMappings := strings.Split(portStr, ",")

	for _, mapping := range portMappings {
		mapping = strings.TrimSpace(mapping)
		if mapping == "" {
			continue
		}

		port := models.ContainerPort{
			Type: "tcp", // default
		}

		// Parse formats like:
		// "0.0.0.0:3000->3000/tcp"
		// "3000/tcp"
		// "127.0.0.1:5432->5432/tcp"

		if strings.Contains(mapping, "->") {
			parts := strings.Split(mapping, "->")
			if len(parts) == 2 {
				// External mapping
				left := parts[0]
				right := parts[1]

				// Parse right side (container port)
				rightParts := strings.Split(right, "/")
				if containerPort, err := strconv.ParseUint(rightParts[0], 10, 16); err == nil {
					port.PrivatePort = uint16(containerPort)
				}
				if len(rightParts) > 1 {
					port.Type = rightParts[1]
				}

				// Parse left side (host binding)
				if strings.Contains(left, ":") {
					hostParts := strings.Split(left, ":")
					if len(hostParts) == 2 {
						port.IP = hostParts[0]
						if hostPort, err := strconv.ParseUint(hostParts[1], 10, 16); err == nil {
							port.PublicPort = uint16(hostPort)
						}
					}
				}
			}
		} else {
			// Just container port
			parts := strings.Split(mapping, "/")
			if containerPort, err := strconv.ParseUint(parts[0], 10, 16); err == nil {
				port.PrivatePort = uint16(containerPort)
			}
			if len(parts) > 1 {
				port.Type = parts[1]
			}
		}

		ports = append(ports, port)
	}

	return ports
}

// parseMemoryUsage extracts memory usage in MB
func (ds *DockerService) parseMemoryUsage(memStr string) float64 {
	// Format: "100MiB / 1GiB" - we want the first part
	parts := strings.Split(memStr, "/")
	if len(parts) == 0 {
		return 0
	}

	usage := strings.TrimSpace(parts[0])

	// Remove units and convert to MB
	if strings.HasSuffix(usage, "GiB") {
		if val, err := strconv.ParseFloat(strings.Replace(usage, "GiB", "", -1), 64); err == nil {
			return val * 1024 // Convert GiB to MiB
		}
	} else if strings.HasSuffix(usage, "MiB") {
		if val, err := strconv.ParseFloat(strings.Replace(usage, "MiB", "", -1), 64); err == nil {
			return val
		}
	} else if strings.HasSuffix(usage, "KiB") {
		if val, err := strconv.ParseFloat(strings.Replace(usage, "KiB", "", -1), 64); err == nil {
			return val / 1024 // Convert KiB to MiB
		}
	}

	return 0
}

// parseNetworkIO extracts network RX/TX bytes
func (ds *DockerService) parseNetworkIO(netStr string) (uint64, uint64) {
	// Format: "1.2kB / 2.3kB"
	parts := strings.Split(netStr, "/")
	if len(parts) != 2 {
		return 0, 0
	}

	rx := ds.parseNetworkValue(strings.TrimSpace(parts[0]))
	tx := ds.parseNetworkValue(strings.TrimSpace(parts[1]))

	return rx, tx
}

// parseNetworkValue converts network values to bytes
func (ds *DockerService) parseNetworkValue(val string) uint64 {
	val = strings.TrimSpace(val)

	multiplier := uint64(1)
	if strings.HasSuffix(val, "kB") {
		multiplier = 1000
		val = strings.Replace(val, "kB", "", -1)
	} else if strings.HasSuffix(val, "MB") {
		multiplier = 1000000
		val = strings.Replace(val, "MB", "", -1)
	} else if strings.HasSuffix(val, "GB") {
		multiplier = 1000000000
		val = strings.Replace(val, "GB", "", -1)
	}

	if result, err := strconv.ParseFloat(val, 64); err == nil {
		return uint64(result * float64(multiplier))
	}

	return 0
}

// runCommand executes a command with timeout
// nolint:unused
func (ds *DockerService) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	// Set timeout to prevent hanging
	done := make(chan error, 1)
	var output []byte
	var err error

	go func() {
		output, err = cmd.Output()
		done <- err
	}()

	select {
	case <-done:
		if err != nil {
			return "", fmt.Errorf("docker command failed: %v", err)
		}
		return string(output), nil
	case <-time.After(10 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return "", fmt.Errorf("docker command timed out after 10 seconds")
	}
}

// runCommandSimple executes a command without capturing output
func (ds *DockerService) runCommandSimple(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// Set timeout to prevent hanging
	done := make(chan error, 1)

	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return fmt.Errorf("docker command timed out after 10 seconds")
	}
}

// StartContainer starts a stopped container
func (ds *DockerService) StartContainer(containerID string) error {
	if !ds.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}

	err := ds.runCommandSimple("docker", "start", containerID)
	if err != nil {
		ds.clearCache()
		return fmt.Errorf("failed to start container: %v", err)
	}

	ds.clearCache()
	return nil
}

// StopContainer stops a running container
func (ds *DockerService) StopContainer(containerID string) error {
	if !ds.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}

	err := ds.runCommandSimple("docker", "stop", "-t", "10", containerID)
	if err != nil {
		ds.clearCache()
		return fmt.Errorf("failed to stop container: %v", err)
	}

	ds.clearCache()
	return nil
}

// RestartContainer restarts a container
func (ds *DockerService) RestartContainer(containerID string) error {
	if !ds.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}

	err := ds.runCommandSimple("docker", "restart", "-t", "10", containerID)
	if err != nil {
		ds.clearCache()
		return fmt.Errorf("failed to restart container: %v", err)
	}

	ds.clearCache()
	return nil
}

// RemoveContainer removes a stopped container
func (ds *DockerService) RemoveContainer(containerID string) error {
	if !ds.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}

	err := ds.runCommandSimple("docker", "rm", containerID)
	if err != nil {
		ds.clearCache()
		return fmt.Errorf("failed to remove container: %v", err)
	}

	ds.clearCache()
	return nil
}
