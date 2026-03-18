# SysMind

<div align="center">

![SysMind Logo](https://via.placeholder.com/200x80/7aa2f7/ffffff?text=SysMind)

**AI-powered system monitoring assistant that helps you understand what your computer is doing in real-time.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org/)
[![Wails](https://img.shields.io/badge/built%20with-Wails-red)](https://wails.io/)
[![React](https://img.shields.io/badge/frontend-React-blue)](https://reactjs.org/)

[Features](#features) •
[Installation](#installation) •
[Usage](#usage) •
[Contributing](#contributing) •
[License](#license)

</div>

## Overview

SysMind is a cross-platform desktop application built with Go and Wails that provides intelligent system monitoring with AI-powered insights. Get real-time information about your system processes, network activity, and resource usage, then ask natural language questions to understand what's happening.

## ✨ Features

### 🖥️ **System Dashboard**
- **Process Monitoring**: View all running processes with CPU and memory usage
- **Network Analysis**: Monitor open ports (TCP/UDP) with associated processes  
- **Resource Timeline**: Real-time charts showing CPU, memory, disk, and network usage
- **Bandwidth Tracking**: Network usage per application with upload/download speeds
- **Auto-refresh**: Live updates every 3-5 seconds

### 🤖 **AI-Powered Insights**
- **Natural Language Queries**: Ask questions about your system in plain English
- **Smart Analysis**: Get explanations about system activity and performance
- **Security Assessment**: Identify potential issues or suspicious behavior
- **Performance Optimization**: Recommendations for system improvements

### 🔌 **AI Provider Support**
- **OpenAI**: GPT-4, GPT-4 Turbo, GPT-3.5 Turbo
- **Cloudflare Workers AI**: Llama 3.1, Llama 3, Mistral 7B
- **Local LLM**: Ollama support (Llama 3, Mistral, Code Llama, etc.)
- **Flexible Configuration**: Easy switching between providers

### 🌍 **Cross-Platform**
- **Linux**: Native `/proc` filesystem integration
- **macOS**: Uses `lsof` and system APIs  
- **Windows**: Leverages `netstat` and WMI

## 📋 Prerequisites

- **Go**: 1.21 or higher
- **Node.js**: 18 or higher  
- **Wails CLI**: v2 (latest)
- **Ollama**: Optional, for local AI models

## 🚀 Installation

### Pre-built Releases (Recommended)

Download the latest release for your platform from the [Releases](https://github.com/Anuolu-2020/sysmind/releases) page:

#### Linux
```bash
# Download and extract
wget https://github.com/Anuolu-2020/sysmind/releases/latest/download/sysmind-v0.1.0-linux-amd64.tar.gz
tar -xzf sysmind-v0.1.0-linux-amd64.tar.gz
chmod +x sysmind
./sysmind
```

#### macOS
```bash
# Intel Macs
wget https://github.com/Anuolu-2020/sysmind/releases/latest/download/sysmind-v0.1.0-darwin-amd64.tar.gz
tar -xzf sysmind-v0.1.0-darwin-amd64.tar.gz

# Apple Silicon Macs  
wget https://github.com/Anuolu-2020/sysmind/releases/latest/download/sysmind-v0.1.0-darwin-arm64.tar.gz
tar -xzf sysmind-v0.1.0-darwin-arm64.tar.gz

chmod +x sysmind && ./sysmind
```

#### Windows
1. Download `sysmind-v0.1.0-windows-amd64.zip`
2. Extract the ZIP file
3. Run `sysmind.exe`

### Build from Source

For developers or custom builds:

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone the repository
git clone https://github.com/Anuolu-2020/sysmind.git
cd sysmind

# Quick setup
make setup

# Development mode
make dev

# Production build
make build
```

## ⚙️ Configuration

### OpenAI Setup
1. Get your API key from [OpenAI Platform](https://platform.openai.com/api-keys)
2. Open SysMind → Settings
3. Select "OpenAI" as provider
4. Enter your API key and choose model (GPT-4 recommended)

### Cloudflare Workers AI Setup  
1. Get API token from [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
2. Find your Account ID in the dashboard
3. Select "Cloudflare Workers AI" in Settings
4. Enter API token, Account ID, and choose model

### Local LLM Setup (Ollama)
1. Install [Ollama](https://ollama.ai)
2. Pull a model: `ollama pull llama3.1`
3. Start server: `ollama serve`  
4. Select "Local LLM" in Settings (endpoint: `http://localhost:11434`)

## 💡 Usage

### Example Queries
- *"Why is my computer running slow?"*
- *"What processes are using the most CPU?"*
- *"Is anything suspicious running on my system?"*
- *"What's using my internet connection?"*
- *"Show me all listening network ports"*
- *"Explain what this process does"*
- *"How can I optimize my system performance?"*

### Dashboard Navigation
- **Processes Tab**: View and sort running processes
- **Ports Tab**: Monitor network connections  
- **Timeline Tab**: Historical resource usage charts
- **Chat Tab**: AI assistant for system queries

## 🏗️ Project Structure

```
sysmind/
├── main.go                    # Application entry point
├── app.go                     # Main app service with Wails bindings  
├── wails.json                 # Wails configuration
├── go.mod & go.sum           # Go dependencies
├── internal/
│   ├── models/               # Data structures and types
│   ├── collectors/           # Platform-specific system data collectors
│   │   ├── linux.go         # Linux implementation
│   │   ├── darwin.go        # macOS implementation  
│   │   └── windows.go       # Windows implementation
│   ├── ai/                   # AI provider implementations
│   │   ├── openai.go        # OpenAI integration
│   │   ├── cloudflare.go    # Cloudflare Workers AI
│   │   └── ollama.go        # Local LLM support
│   └── services/             # Business logic and configuration
└── frontend/
    ├── src/
    │   ├── App.jsx           # Main React component
    │   ├── components/       # UI components
    │   │   ├── Dashboard.jsx # Main dashboard  
    │   │   ├── ProcessList.jsx
    │   │   ├── PortList.jsx
    │   │   ├── ResourceTimeline.jsx
    │   │   └── ChatPanel.jsx
    │   └── style.css         # Application styles
    └── package.json          # Frontend dependencies
```

## 🔨 Building for Distribution

```bash
# Build for current platform
make build

# Build for all platforms
make cross-build

# Development commands
make dev          # Start development server
make test         # Run tests
make lint         # Run linters
make clean        # Clean build artifacts

# Release management
make version-check          # Check current version
make release VERSION=1.2.3  # Create new release
```

For detailed build instructions, see [RELEASE.md](RELEASE.md).

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Quick Development Setup
```bash
# Fork and clone the repository
git clone https://github.com/Anuolu-2020/sysmind.git
cd sysmind

# Setup development environment  
make setup

# Start development
make dev
```

### Release Process
For maintainers releasing new versions:

```bash
# Check what's changed since last release
make version-check

# Create a new release (automated)
make release VERSION=1.2.3
```

See [RELEASE.md](RELEASE.md) for complete release documentation.

### Areas for Contribution
- 🐛 Bug fixes and performance improvements
- ✨ New features and AI integrations
- 📖 Documentation improvements  
- 🌍 Internationalization/translations
- 🧪 Additional test coverage
- 📦 Package managers and distribution

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Wails](https://wails.io/) - Modern desktop app framework
- [React](https://reactjs.org/) - Frontend UI library
- [gopsutil](https://github.com/shirou/gopsutil) - Cross-platform system information
- [Ollama](https://ollama.ai/) - Local LLM runtime

## 📞 Support

- 🐛 **Bug Reports**: [Open an issue](https://github.com/Anuolu-2020/sysmind/issues)
- 💡 **Feature Requests**: [Discussions](https://github.com/Anuolu-2020/sysmind/discussions)  
- 📧 **Contact**: [anuolu2020@gmail.com](mailto:anuolu2020@gmail.com)

---

<div align="center">

**⭐ If you find SysMind useful, please consider giving it a star!**

Made with ❤️ by the open source community

</div>
