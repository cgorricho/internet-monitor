# 📡 Internet Connection Monitor

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey.svg)](#installation)

A high-performance Go application that monitors internet connection performance with on-demand dashboard generation and optional machine pairing for comparative analysis.

## 🎯 Key Features

- **🚀 Single Binary Deployment** - Zero dependencies, runs anywhere
- **📊 On-Demand Dashboards** - Generate static HTML reports when needed
- **🔗 Machine Pairing** - Compare performance across multiple locations
- **⚡ Resource Efficient** - 80% less memory usage than dashboard servers
- **🛡️ Legal Compliance** - No Ookla dependency, uses open-source speed tests
- **🔧 Smart Scheduling** - Peak/off-peak intelligent monitoring
- **📱 Mobile Responsive** - Dashboards work perfectly on mobile devices

## 📊 Monitored Metrics

### Core Performance
- **Download/Upload Speed** (Mbps) - Via HTTP file tests or LibreSpeed
- **Ping Latency** (ms) - Network responsiveness
- **DNS Resolution** (ms) - Web browsing performance
- **WiFi Signal Strength** (dBm) - Local connection quality
- **Public IP Tracking** - ISP connection stability

### Alert System
- Configurable warning and critical thresholds
- Automatic alert generation and tracking
- Historical alert analysis

## 🚀 Quick Start

### Option 1: Download Pre-built Binary

```bash
# Download for your platform
wget https://github.com/cgorricho/internet-monitor/releases/latest/download/internet-monitor-linux-amd64
chmod +x internet-monitor-linux-amd64
sudo mv internet-monitor-linux-amd64 /usr/local/bin/internet-monitor
```

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/cgorricho/internet-monitor.git
cd internet-monitor

# Build binary
./scripts/build.sh

# Copy to system PATH
sudo cp dist/internet-monitor-linux-amd64 /usr/local/bin/internet-monitor
```

### Initialize and Start

```bash
# Initialize database
internet-monitor init

# Start monitoring (runs in background)
internet-monitor monitor &

# Generate dashboard
./scripts/dashboard.sh
```

## 🖥️ Dashboard Generation

Generate beautiful, responsive HTML dashboards on-demand:

```bash
# Basic 24-hour dashboard
./scripts/dashboard.sh

# Weekly performance report
./scripts/dashboard.sh --hours 168 --output weekly-report.html

# Comparative dashboard (when paired with another machine)
./scripts/dashboard.sh --compare

# Generate without opening browser
./scripts/dashboard.sh --no-browser
```

### Dashboard Features
- 📈 **Time-series charts** with ApexCharts (140KB vs 3MB Plotly)
- 📱 **Mobile responsive** design
- 🌐 **Offline viewing** - works without internet
- 📤 **Shareable reports** - send HTML files via email
- ⚡ **Fast loading** - static files load instantly

## 🔗 Machine Pairing

Compare internet performance between locations (e.g., home office vs main office):

```bash
# On first machine (e.g., office server)
internet-monitor serve  # Start API server
internet-monitor pair   # Generate pairing code

# On second machine (e.g., home laptop)
internet-monitor pair ABC123  # Enter pairing code

# Generate comparative dashboard on either machine
./scripts/dashboard.sh --compare
```

### Pairing Benefits
- 🏠 **Home vs Office** comparison
- 📊 **Side-by-side performance** graphs
- 🔐 **Secure authentication** with TLS
- 🌐 **Resilient to network issues** with periodic sync

## ⚙️ Configuration

Configuration via YAML file or environment variables:

```yaml
# config.yaml
database:
  path: "./data/internet-monitor.db"
  retention_days: 90

monitor:
  peak_hours:
    start: "09:00"
    end: "18:00"
  intervals:
    peak:
      quick: "2m"
      speed_test: "10m"
    off_peak:
      quick: "5m"
      speed_test: "20m"

network:
  speed_test:
    primary: "httpfile"
    backends:
      httpfile:
        options:
          download_url: "https://proof.ovh.net/files/100Mb.dat"
```

## 🛠️ Available Commands

| Command | Description |
|---------|-------------|
| `internet-monitor init` | Initialize database and configuration |
| `internet-monitor monitor` | Start background monitoring service |
| `internet-monitor dashboard` | Generate static HTML dashboard |
| `internet-monitor serve` | Start API server for pairing |
| `internet-monitor pair [code]` | Generate or join pairing |
| `internet-monitor status` | Show system status and statistics |
| `internet-monitor --help` | Show all available options |

## 📦 Installation

### Linux (Ubuntu/Debian)

```bash
# Install system dependencies
sudo apt update && sudo apt install -y curl wget net-tools dnsutils wireless-tools

# Install binary
wget -O internet-monitor https://github.com/cgorricho/internet-monitor/releases/latest/download/internet-monitor-linux-amd64
chmod +x internet-monitor
sudo mv internet-monitor /usr/local/bin/
```

### Windows

1. Download `internet-monitor-windows-amd64.exe` from [releases](https://github.com/cgorricho/internet-monitor/releases)
2. Run the MSI installer (coming soon) or:
3. Place the `.exe` file in your PATH
4. Open Command Prompt as Administrator and run `internet-monitor init`

### macOS

```bash
# Using Homebrew (coming soon)
brew install cgorricho/tap/internet-monitor

# Or download directly
curl -L -o internet-monitor https://github.com/cgorricho/internet-monitor/releases/latest/download/internet-monitor-darwin-amd64
chmod +x internet-monitor
sudo mv internet-monitor /usr/local/bin/
```

## 🔧 System Service Setup

### Linux (systemd)

```bash
# Create service file
sudo tee /etc/systemd/system/internet-monitor.service > /dev/null <<EOF
[Unit]
Description=Internet Monitor
After=network.target

[Service]
Type=simple
User=monitor
WorkingDirectory=/opt/internet-monitor
ExecStart=/usr/local/bin/internet-monitor monitor
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl enable internet-monitor
sudo systemctl start internet-monitor
```

### Windows Service

```powershell
# Install as Windows Service (PowerShell as Administrator)
internet-monitor install-service
net start "Internet Monitor"
```

## 📊 Performance Comparison

| Feature | Previous (Python/Plotly) | New (Go/Static) | Improvement |
|---------|-------------------------|-----------------|-------------|
| **Memory Usage** | 50-100MB | 10-20MB | 60-80% reduction |
| **Startup Time** | 10+ seconds | <2 seconds | 80% faster |
| **Dashboard Size** | 3MB+ Plotly.js | 140KB ApexCharts | 95% smaller |
| **Binary Size** | 100MB+ (PyInstaller) | ~15MB | 85% smaller |
| **Resource Usage** | High (continuous server) | Minimal (on-demand) | 90% reduction |
| **Deployment** | Complex (Python + deps) | Single binary | Zero dependencies |

## 🚨 Troubleshooting

### Common Issues

**Dashboard generation fails:**
```bash
# Check if binary is accessible
internet-monitor status

# Verify database exists
ls -la data/internet-monitor.db

# Check logs
tail -f logs/internet-monitor.log
```

**Network tests failing:**
```bash
# Test basic connectivity
ping 8.8.8.8

# Check DNS resolution
nslookup google.com

# Verify WiFi tools (Linux)
iwconfig
```

**Pairing issues:**
```bash
# Check if API server is running
internet-monitor serve &

# Verify firewall settings
sudo ufw allow 8080
```

### Log Locations
- **Linux**: `./logs/internet-monitor.log`
- **Windows**: `%APPDATA%\internet-monitor\logs\`
- **macOS**: `~/Library/Logs/internet-monitor/`

## 🧪 Development

### Building from Source

```bash
# Install Go 1.21+
sudo apt install golang-go

# Clone and build
git clone https://github.com/cgorricho/internet-monitor.git
cd internet-monitor

# Download dependencies
go mod tidy

# Build for all platforms
./scripts/build.sh

# Run tests
go test ./...
```

### Project Structure

```
internet-monitor/
├── cmd/internet-monitor/     # Main application
├── internal/                 # Internal packages
│   ├── config/              # Configuration management
│   ├── database/            # SQLite operations
│   ├── network/             # Speed testing backends
│   ├── monitor/             # Background monitoring
│   ├── dashboard/           # HTML generation
│   └── server/              # API server for pairing
├── web/                     # Dashboard templates and assets
├── scripts/                 # Build and deployment scripts
├── configs/                 # Configuration examples
└── old_files/               # Legacy Python implementation
```

## 📚 Documentation

- [Architecture Overview](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Configuration Reference](docs/configuration.md)
- [Deployment Guide](docs/deployment.md)
- [Contributing Guidelines](CONTRIBUTING.md)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **LibreSpeed** for open-source speed testing
- **ApexCharts** for lightweight, beautiful charts
- **Go community** for excellent tooling and libraries

---

**Created by**: Carlos Gorricho  
**Last Updated**: October 2024  
**Version**: 2.0.0 (Go Rewrite)

[![Star this project](https://img.shields.io/github/stars/cgorricho/internet-monitor?style=social)](https://github.com/cgorricho/internet-monitor)
