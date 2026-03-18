package services

import (
	"fmt"
	"strings"

	"sysmind/internal/models"
)

// DevEnvironmentService detects and identifies development environments
type DevEnvironmentService struct {
	dockerService *DockerService
}

// NewDevEnvironmentService creates a new dev environment service
func NewDevEnvironmentService(dockerService *DockerService) *DevEnvironmentService {
	return &DevEnvironmentService{
		dockerService: dockerService,
	}
}

// GetDevEnvironmentInfo retrieves all development environment information
func (des *DevEnvironmentService) GetDevEnvironmentInfo(processes []models.ProcessInfo, ports []models.PortInfo) (*models.DevEnvironmentInfo, error) {
	info := &models.DevEnvironmentInfo{
		Containers:   []models.DockerContainer{},
		Environments: []models.DevEnvironment{},
		DevPorts:     []models.DevPort{},
	}

	// Get Docker containers if available
	if des.dockerService != nil {
		info.DockerRunning = des.dockerService.IsDockerAvailable()
		if info.DockerRunning {
			containers, err := des.dockerService.GetContainers()
			if err == nil {
				info.Containers = containers
			}
		}
	}

	// Detect development ports
	info.DevPorts = des.identifyDevPorts(ports, processes)

	// Create development environments from containers and processes
	info.Environments = des.createDevEnvironments(info.Containers, info.DevPorts, processes)

	return info, nil
}

// identifyDevPorts identifies development-related ports with intelligent framework detection
func (des *DevEnvironmentService) identifyDevPorts(ports []models.PortInfo, processes []models.ProcessInfo) []models.DevPort {
	var devPorts []models.DevPort

	// Create process lookup map
	processMap := make(map[int32]models.ProcessInfo)
	for _, proc := range processes {
		processMap[proc.PID] = proc
	}

	for _, port := range ports {
		if port.State != "LISTENING" {
			continue
		}

		devPort := des.analyzePort(port, processMap)
		if devPort.Technology != "" {
			devPorts = append(devPorts, devPort)
		}
	}

	return devPorts
}

// analyzePort analyzes a port to determine if it's a development service
func (des *DevEnvironmentService) analyzePort(port models.PortInfo, processMap map[int32]models.ProcessInfo) models.DevPort {
	devPort := models.DevPort{
		Port:        uint16(port.Port),
		ProcessName: port.ProcessName,
		ProcessPID:  port.PID,
	}

	process, hasProcess := processMap[port.PID]

	// Analyze by port number first
	devPort = des.analyzeByPortNumber(devPort)

	// Analyze by process name and command if we have process info
	if hasProcess {
		devPort = des.analyzeByProcess(devPort, process)
	}

	// Generate URL if it's a web service
	if devPort.Technology != "" && des.isWebService(devPort.Technology) {
		devPort.URL = fmt.Sprintf("http://localhost:%d", devPort.Port)
	}

	return devPort
}

// analyzeByPortNumber identifies services by common development port numbers
func (des *DevEnvironmentService) analyzeByPortNumber(devPort models.DevPort) models.DevPort {
	port := devPort.Port

	switch port {
	case 3000:
		devPort.Technology = "nextjs"
		devPort.Framework = "Next.js / React"
		devPort.Icon = "⚛️"
		devPort.Description = "Next.js development server or React app"
	case 3001:
		devPort.Technology = "react"
		devPort.Framework = "React Dev Server"
		devPort.Icon = "⚛️"
		devPort.Description = "React development server (alternative port)"
	case 4000:
		devPort.Technology = "vue"
		devPort.Framework = "Vue.js / Nuxt.js"
		devPort.Icon = "🟢"
		devPort.Description = "Vue.js or Nuxt.js development server"
	case 4200:
		devPort.Technology = "angular"
		devPort.Framework = "Angular"
		devPort.Icon = "🅰️"
		devPort.Description = "Angular development server"
	case 5173, 5174:
		devPort.Technology = "vite"
		devPort.Framework = "Vite"
		devPort.Icon = "⚡"
		devPort.Description = "Vite development server"
	case 8080:
		devPort.Technology = "web"
		devPort.Framework = "Web Server"
		devPort.Icon = "🌐"
		devPort.Description = "Generic web server (often Spring Boot, Tomcat)"
	case 8000:
		devPort.Technology = "web"
		devPort.Framework = "Web Server"
		devPort.Icon = "🌐"
		devPort.Description = "Web server (often Django, Python HTTP server)"
	case 8888:
		devPort.Technology = "jupyter"
		devPort.Framework = "Jupyter"
		devPort.Icon = "📓"
		devPort.Description = "Jupyter Notebook server"
	case 9000:
		devPort.Technology = "web"
		devPort.Framework = "Web Server"
		devPort.Icon = "🌐"
		devPort.Description = "Web application server"
	case 5432:
		devPort.Technology = "postgres"
		devPort.Framework = "PostgreSQL"
		devPort.Icon = "🐘"
		devPort.Description = "PostgreSQL database server"
	case 3306:
		devPort.Technology = "mysql"
		devPort.Framework = "MySQL"
		devPort.Icon = "🐬"
		devPort.Description = "MySQL database server"
	case 6379:
		devPort.Technology = "redis"
		devPort.Framework = "Redis"
		devPort.Icon = "🔴"
		devPort.Description = "Redis in-memory database"
	case 27017:
		devPort.Technology = "mongodb"
		devPort.Framework = "MongoDB"
		devPort.Icon = "🍃"
		devPort.Description = "MongoDB NoSQL database"
	case 5000:
		devPort.Technology = "flask"
		devPort.Framework = "Flask / Python"
		devPort.Icon = "🐍"
		devPort.Description = "Flask web application or Python HTTP server"
	case 1337, 1338:
		devPort.Technology = "strapi"
		devPort.Framework = "Strapi"
		devPort.Icon = "🚀"
		devPort.Description = "Strapi headless CMS"
	case 11211:
		devPort.Technology = "memcached"
		devPort.Framework = "Memcached"
		devPort.Icon = "💾"
		devPort.Description = "Memcached memory caching system"
	case 9200:
		devPort.Technology = "elasticsearch"
		devPort.Framework = "Elasticsearch"
		devPort.Icon = "🔍"
		devPort.Description = "Elasticsearch search engine"
	case 5672:
		devPort.Technology = "rabbitmq"
		devPort.Framework = "RabbitMQ"
		devPort.Icon = "🐰"
		devPort.Description = "RabbitMQ message broker"
	case 9092:
		devPort.Technology = "kafka"
		devPort.Framework = "Apache Kafka"
		devPort.Icon = "📡"
		devPort.Description = "Apache Kafka message streaming"
	case 2181:
		devPort.Technology = "zookeeper"
		devPort.Framework = "Apache ZooKeeper"
		devPort.Icon = "🦓"
		devPort.Description = "Apache ZooKeeper coordination service"
	case 5601:
		devPort.Technology = "kibana"
		devPort.Framework = "Kibana"
		devPort.Icon = "📊"
		devPort.Description = "Kibana visualization dashboard"
	case 3002:
		devPort.Technology = "storybook"
		devPort.Framework = "Storybook"
		devPort.Icon = "📚"
		devPort.Description = "Storybook component development"
	case 6006:
		devPort.Technology = "storybook"
		devPort.Framework = "Storybook"
		devPort.Icon = "📚"
		devPort.Description = "Storybook component development"
	case 24678:
		devPort.Technology = "vite"
		devPort.Framework = "Vite HMR"
		devPort.Icon = "⚡"
		devPort.Description = "Vite Hot Module Replacement server"
	case 35729:
		devPort.Technology = "livereload"
		devPort.Framework = "LiveReload"
		devPort.Icon = "🔄"
		devPort.Description = "LiveReload development server"
	case 7000, 7001:
		devPort.Technology = "gatsby"
		devPort.Framework = "Gatsby"
		devPort.Icon = "⚡"
		devPort.Description = "Gatsby static site generator"
	case 4444:
		devPort.Technology = "selenium"
		devPort.Framework = "Selenium Grid"
		devPort.Icon = "🕷️"
		devPort.Description = "Selenium WebDriver server"
	case 9229:
		devPort.Technology = "nodejs-debug"
		devPort.Framework = "Node.js Inspector"
		devPort.Icon = "🐛"
		devPort.Description = "Node.js debugging inspector"
	case 5858:
		devPort.Technology = "nodejs-debug"
		devPort.Framework = "Node.js Debug"
		devPort.Icon = "🐛"
		devPort.Description = "Node.js debug server (legacy)"
	case 2049:
		devPort.Technology = "nfs"
		devPort.Framework = "NFS"
		devPort.Icon = "📁"
		devPort.Description = "Network File System"
	case 8081, 8082, 8083, 8084, 8085:
		devPort.Technology = "web-alt"
		devPort.Framework = "Alternative Web Server"
		devPort.Icon = "🌐"
		devPort.Description = "Alternative web server port"
	case 3003, 3004, 3005:
		devPort.Technology = "react-alt"
		devPort.Framework = "React/Next.js Alt"
		devPort.Icon = "⚛️"
		devPort.Description = "React or Next.js on alternative port"
	case 8787:
		devPort.Technology = "rstudio"
		devPort.Framework = "RStudio Server"
		devPort.Icon = "📈"
		devPort.Description = "RStudio Server for R development"
	case 9090:
		devPort.Technology = "prometheus"
		devPort.Framework = "Prometheus"
		devPort.Icon = "🔥"
		devPort.Description = "Prometheus monitoring server"
	case 3300:
		devPort.Technology = "grafana"
		devPort.Framework = "Grafana"
		devPort.Icon = "📈"
		devPort.Description = "Grafana monitoring dashboard"
	case 8086:
		devPort.Technology = "influxdb"
		devPort.Framework = "InfluxDB"
		devPort.Icon = "📊"
		devPort.Description = "InfluxDB time series database"
	case 8500:
		devPort.Technology = "consul"
		devPort.Framework = "HashiCorp Consul"
		devPort.Icon = "🔗"
		devPort.Description = "Consul service discovery"
	case 8200:
		devPort.Technology = "vault"
		devPort.Framework = "HashiCorp Vault"
		devPort.Icon = "🔐"
		devPort.Description = "Vault secret management"
	case 4040:
		devPort.Technology = "spark"
		devPort.Framework = "Apache Spark"
		devPort.Icon = "⚡"
		devPort.Description = "Apache Spark web UI"
	}

	return devPort
}

// analyzeByProcess analyzes process information to identify frameworks
func (des *DevEnvironmentService) analyzeByProcess(devPort models.DevPort, process models.ProcessInfo) models.DevPort {
	cmdLine := strings.ToLower(process.CommandLine)
	processName := strings.ToLower(process.Name)

	// Next.js detection
	if strings.Contains(cmdLine, "next") || strings.Contains(cmdLine, "@next/") {
		devPort.Technology = "nextjs"
		devPort.Framework = "Next.js"
		devPort.Icon = "⚛️"
		devPort.Description = "Next.js React framework"
	} else if strings.Contains(cmdLine, "react-scripts") || strings.Contains(cmdLine, "create-react-app") {
		devPort.Technology = "react"
		devPort.Framework = "Create React App"
		devPort.Icon = "⚛️"
		devPort.Description = "Create React App development server"
	} else if strings.Contains(cmdLine, "vue-cli-service") || strings.Contains(cmdLine, "@vue/cli") {
		devPort.Technology = "vue"
		devPort.Framework = "Vue CLI"
		devPort.Icon = "🟢"
		devPort.Description = "Vue.js CLI development server"
	} else if strings.Contains(cmdLine, "nuxt") {
		devPort.Technology = "nuxt"
		devPort.Framework = "Nuxt.js"
		devPort.Icon = "💚"
		devPort.Description = "Nuxt.js Vue.js framework"
	} else if strings.Contains(cmdLine, "vite") || strings.Contains(processName, "vite") {
		devPort.Technology = "vite"
		devPort.Framework = "Vite"
		devPort.Icon = "⚡"
		devPort.Description = "Vite build tool and dev server"
	} else if strings.Contains(cmdLine, "webpack") || strings.Contains(cmdLine, "webpack-dev-server") {
		devPort.Technology = "webpack"
		devPort.Framework = "Webpack Dev Server"
		devPort.Icon = "📦"
		devPort.Description = "Webpack development server"
	} else if strings.Contains(cmdLine, "nodemon") {
		devPort.Technology = "nodejs"
		devPort.Framework = "Node.js (Nodemon)"
		devPort.Icon = "🟢"
		devPort.Description = "Node.js application with auto-restart"
	} else if strings.Contains(cmdLine, "pm2") {
		devPort.Technology = "nodejs"
		devPort.Framework = "Node.js (PM2)"
		devPort.Icon = "🟢"
		devPort.Description = "Node.js application managed by PM2"
	} else if strings.Contains(cmdLine, "gatsby") {
		devPort.Technology = "gatsby"
		devPort.Framework = "Gatsby"
		devPort.Icon = "⚡"
		devPort.Description = "Gatsby static site generator"
	} else if strings.Contains(cmdLine, "storybook") {
		devPort.Technology = "storybook"
		devPort.Framework = "Storybook"
		devPort.Icon = "📚"
		devPort.Description = "Storybook component development"
	} else if strings.Contains(cmdLine, "snowpack") {
		devPort.Technology = "snowpack"
		devPort.Framework = "Snowpack"
		devPort.Icon = "❄️"
		devPort.Description = "Snowpack frontend build tool"
	} else if strings.Contains(cmdLine, "parcel") {
		devPort.Technology = "parcel"
		devPort.Framework = "Parcel"
		devPort.Icon = "📦"
		devPort.Description = "Parcel bundler development server"
	} else if strings.Contains(cmdLine, "rollup") {
		devPort.Technology = "rollup"
		devPort.Framework = "Rollup"
		devPort.Icon = "📦"
		devPort.Description = "Rollup module bundler"
	} else if strings.Contains(cmdLine, "svelte") || strings.Contains(cmdLine, "sveltekit") {
		devPort.Technology = "svelte"
		devPort.Framework = "Svelte/SvelteKit"
		devPort.Icon = "🔥"
		devPort.Description = "Svelte frontend framework"
	} else if strings.Contains(cmdLine, "astro") {
		devPort.Technology = "astro"
		devPort.Framework = "Astro"
		devPort.Icon = "🚀"
		devPort.Description = "Astro static site generator"
	} else if strings.Contains(cmdLine, "remix") {
		devPort.Technology = "remix"
		devPort.Framework = "Remix"
		devPort.Icon = "💿"
		devPort.Description = "Remix React framework"
	} else if strings.Contains(processName, "node") && strings.Contains(cmdLine, "server") {
		devPort.Technology = "nodejs"
		devPort.Framework = "Node.js"
		devPort.Icon = "🟢"
		devPort.Description = "Node.js web server"
	} else if strings.Contains(processName, "deno") {
		devPort.Technology = "deno"
		devPort.Framework = "Deno"
		devPort.Icon = "🦕"
		devPort.Description = "Deno JavaScript runtime"
	} else if strings.Contains(processName, "bun") {
		devPort.Technology = "bun"
		devPort.Framework = "Bun"
		devPort.Icon = "🍞"
		devPort.Description = "Bun JavaScript runtime"
	} else if strings.Contains(processName, "python") {
		if strings.Contains(cmdLine, "django") {
			devPort.Technology = "django"
			devPort.Framework = "Django"
			devPort.Icon = "🐍"
			devPort.Description = "Django Python web framework"
		} else if strings.Contains(cmdLine, "flask") {
			devPort.Technology = "flask"
			devPort.Framework = "Flask"
			devPort.Icon = "🐍"
			devPort.Description = "Flask Python web framework"
		} else if strings.Contains(cmdLine, "fastapi") {
			devPort.Technology = "fastapi"
			devPort.Framework = "FastAPI"
			devPort.Icon = "🚀"
			devPort.Description = "FastAPI Python web framework"
		} else if strings.Contains(cmdLine, "streamlit") {
			devPort.Technology = "streamlit"
			devPort.Framework = "Streamlit"
			devPort.Icon = "🌊"
			devPort.Description = "Streamlit data app framework"
		} else if strings.Contains(cmdLine, "jupyter") {
			devPort.Technology = "jupyter"
			devPort.Framework = "Jupyter"
			devPort.Icon = "📓"
			devPort.Description = "Jupyter Notebook/Lab server"
		}
	} else if strings.Contains(processName, "go") && (strings.Contains(cmdLine, "run") || strings.Contains(cmdLine, "server")) {
		devPort.Technology = "golang"
		devPort.Framework = "Go"
		devPort.Icon = "🐹"
		devPort.Description = "Go web server or API"
	} else if strings.Contains(processName, "rust") || strings.Contains(cmdLine, "cargo") {
		devPort.Technology = "rust"
		devPort.Framework = "Rust"
		devPort.Icon = "🦀"
		devPort.Description = "Rust web server"
	} else if strings.Contains(processName, "postgres") || strings.Contains(cmdLine, "postgres") {
		devPort.Technology = "postgres"
		devPort.Framework = "PostgreSQL"
		devPort.Icon = "🐘"
		devPort.Description = "PostgreSQL database server"
	} else if strings.Contains(processName, "mysql") || strings.Contains(cmdLine, "mysql") {
		devPort.Technology = "mysql"
		devPort.Framework = "MySQL"
		devPort.Icon = "🐬"
		devPort.Description = "MySQL database server"
	} else if strings.Contains(processName, "redis") || strings.Contains(cmdLine, "redis") {
		devPort.Technology = "redis"
		devPort.Framework = "Redis"
		devPort.Icon = "🔴"
		devPort.Description = "Redis in-memory database"
	} else if strings.Contains(processName, "nginx") {
		devPort.Technology = "nginx"
		devPort.Framework = "Nginx"
		devPort.Icon = "🌐"
		devPort.Description = "Nginx web server"
	} else if strings.Contains(processName, "apache") {
		devPort.Technology = "apache"
		devPort.Framework = "Apache"
		devPort.Icon = "🌐"
		devPort.Description = "Apache HTTP server"
	} else if strings.Contains(cmdLine, "spring-boot") || strings.Contains(cmdLine, "spring.boot") {
		devPort.Technology = "springboot"
		devPort.Framework = "Spring Boot"
		devPort.Icon = "🍃"
		devPort.Description = "Spring Boot Java application"
	} else if strings.Contains(processName, "java") && strings.Contains(cmdLine, "tomcat") {
		devPort.Technology = "tomcat"
		devPort.Framework = "Apache Tomcat"
		devPort.Icon = "🐱"
		devPort.Description = "Apache Tomcat servlet container"
	} else if strings.Contains(cmdLine, "dotnet") || strings.Contains(cmdLine, ".net") {
		devPort.Technology = "dotnet"
		devPort.Framework = ".NET"
		devPort.Icon = "🔵"
		devPort.Description = ".NET application server"
	} else if strings.Contains(processName, "php") {
		devPort.Technology = "php"
		devPort.Framework = "PHP"
		devPort.Icon = "🐘"
		devPort.Description = "PHP web application"
	} else if strings.Contains(cmdLine, "ruby") || strings.Contains(cmdLine, "rails") {
		devPort.Technology = "rails"
		devPort.Framework = "Ruby on Rails"
		devPort.Icon = "💎"
		devPort.Description = "Ruby on Rails web application"
	}

	return devPort
}

// createDevEnvironments creates unified development environment entries
func (des *DevEnvironmentService) createDevEnvironments(containers []models.DockerContainer, devPorts []models.DevPort, processes []models.ProcessInfo) []models.DevEnvironment {
	var environments []models.DevEnvironment

	// Create environments from containers
	for _, container := range containers {
		env := des.containerToEnvironment(container)
		environments = append(environments, env)
	}

	// Create environments from development ports (non-containerized)
	for _, devPort := range devPorts {
		env := des.devPortToEnvironment(devPort, processes)
		environments = append(environments, env)
	}

	return environments
}

// containerToEnvironment converts a Docker container to a development environment
func (des *DevEnvironmentService) containerToEnvironment(container models.DockerContainer) models.DevEnvironment {
	env := models.DevEnvironment{
		ID:          fmt.Sprintf("container-%s", container.ID[:12]),
		ContainerID: container.ID,
		Status:      container.State,
	}

	// Analyze container image to determine technology
	image := strings.ToLower(container.Image)

	if strings.Contains(image, "postgres") {
		env.Name = "PostgreSQL Database"
		env.Type = "database"
		env.Technology = "postgres"
		env.Icon = "🐘"
		env.Description = fmt.Sprintf("PostgreSQL database (%s)", container.Image)
	} else if strings.Contains(image, "redis") {
		env.Name = "Redis Cache"
		env.Type = "database"
		env.Technology = "redis"
		env.Icon = "🔴"
		env.Description = fmt.Sprintf("Redis in-memory store (%s)", container.Image)
	} else if strings.Contains(image, "mongo") {
		env.Name = "MongoDB Database"
		env.Type = "database"
		env.Technology = "mongodb"
		env.Icon = "🍃"
		env.Description = fmt.Sprintf("MongoDB NoSQL database (%s)", container.Image)
	} else if strings.Contains(image, "nginx") {
		env.Name = "Nginx Server"
		env.Type = "proxy"
		env.Technology = "nginx"
		env.Icon = "🌐"
		env.Description = fmt.Sprintf("Nginx web server (%s)", container.Image)
	} else if strings.Contains(image, "node") {
		env.Name = "Node.js App"
		env.Type = "web"
		env.Technology = "nodejs"
		env.Icon = "🟢"
		env.Description = fmt.Sprintf("Node.js application (%s)", container.Image)
	} else if strings.Contains(image, "python") {
		env.Name = "Python App"
		env.Type = "api"
		env.Technology = "python"
		env.Icon = "🐍"
		env.Description = fmt.Sprintf("Python application (%s)", container.Image)
	} else {
		env.Name = container.Name
		env.Type = "container"
		env.Technology = "docker"
		env.Icon = "🐳"
		env.Description = fmt.Sprintf("Docker container (%s)", container.Image)
	}

	// Set ports and URLs
	for _, port := range container.Ports {
		if port.PublicPort > 0 {
			env.Port = port.PublicPort
			if des.isWebService(env.Technology) {
				env.URLs = append(env.URLs, fmt.Sprintf("http://localhost:%d", port.PublicPort))
			}
			break // Use first mapped port
		}
	}

	return env
}

// devPortToEnvironment converts a development port to an environment
func (des *DevEnvironmentService) devPortToEnvironment(devPort models.DevPort, processes []models.ProcessInfo) models.DevEnvironment {
	env := models.DevEnvironment{
		ID:          fmt.Sprintf("port-%d", devPort.Port),
		Name:        devPort.Framework,
		Technology:  devPort.Technology,
		Port:        devPort.Port,
		ProcessName: devPort.ProcessName,
		ProcessPID:  devPort.ProcessPID,
		Status:      "running",
		Icon:        devPort.Icon,
		Description: devPort.Description,
	}

	// Determine type based on technology
	switch devPort.Technology {
	case "postgres", "mysql", "redis", "mongodb":
		env.Type = "database"
	case "nextjs", "react", "vue", "angular":
		env.Type = "web"
	case "nginx", "apache":
		env.Type = "proxy"
	default:
		env.Type = "api"
	}

	// Add URL for web services
	if devPort.URL != "" {
		env.URLs = append(env.URLs, devPort.URL)
	}

	return env
}

// isWebService determines if a technology typically serves web content
func (des *DevEnvironmentService) isWebService(technology string) bool {
	webServices := []string{
		"nextjs", "react", "vue", "angular", "vite", "webpack", "gatsby", "storybook",
		"snowpack", "parcel", "rollup", "svelte", "astro", "remix",
		"nodejs", "deno", "bun", "django", "flask", "fastapi", "streamlit",
		"golang", "rust", "springboot", "tomcat", "dotnet", "php", "rails",
		"web", "web-alt", "nginx", "apache", "react-alt",
	}

	for _, service := range webServices {
		if technology == service {
			return true
		}
	}
	return false
}

// GetContainerActions returns available actions for a container
func (des *DevEnvironmentService) GetContainerActions(containerID string) []string {
	if !des.dockerService.IsDockerAvailable() {
		return []string{}
	}

	// For now, return basic actions - could be enhanced to check actual container status
	return []string{"start", "stop", "restart", "logs", "inspect"}
}
