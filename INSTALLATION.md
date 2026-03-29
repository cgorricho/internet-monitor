# 📋 Internet Monitor - Installation Guide

This guide covers installation and setup for all supported platforms.

## 🚀 Quick Installation

### Linux (Ubuntu/Debian)

```bash
# Install system dependencies
sudo apt update
sudo apt install -y curl wget net-tools dnsutils wireless-tools

# Download and install binary
wget -O internet-monitor https://github.com/cgorricho/internet-monitor/releases/latest/download/internet-monitor-linux-amd64
chmod +x internet-monitor
sudo mv internet-monitor /usr/local/bin/

# Initialize and start
internet-monitor init
internet-monitor monitor &
```

### Windows

1. **Download**: Get `internet-monitor-windows-amd64.exe` from [releases](https://github.com/cgorricho/internet-monitor/releases)
2. **Install**: Place the executable in your PATH or a dedicated folder
3. **Initialize**: Open Command Prompt as Administrator:
   ```cmd
   internet-monitor.exe init
   internet-monitor.exe monitor
   ```

### macOS

```bash
# Install using Homebrew (coming soon)
brew install cgorricho/tap/internet-monitor

# Or download directly
curl -L -o internet-monitor https://github.com/cgorricho/internet-monitor/releases/latest/download/internet-monitor-darwin-amd64
chmod +x internet-monitor
sudo mv internet-monitor /usr/local/bin/

# Initialize and start
internet-monitor init
internet-monitor monitor &
```

## 🏗️ Building from Source

### Prerequisites

- **Go 1.21+**
- **Git**
- **C compiler** (for SQLite CGO)

### Linux/macOS

```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Clone and build
git clone https://github.com/cgorricho/internet-monitor.git
cd internet-monitor
go mod tidy
./scripts/build.sh

# Install binary
sudo cp dist/internet-monitor-linux-amd64 /usr/local/bin/internet-monitor
```

### Windows

```powershell
# Install Go from https://golang.org/dl/
# Install Git from https://git-scm.com/

# Clone and build
git clone https://github.com/cgorricho/internet-monitor.git
cd internet-monitor
go mod tidy
go build -o internet-monitor.exe ./cmd/internet-monitor

# Add to PATH or move to desired location
```

## ⚙️ Configuration

### Basic Setup

```bash
# Copy example configuration
cp configs/config.example.yaml config.yaml

# Edit configuration as needed
nano config.yaml
```

### Key Configuration Options

```yaml
# Essential settings to customize
monitor:
  peak_hours:
    start: "09:00"  # Adjust to your work hours
    end: "18:00"    # Adjust to your work hours
  
  alerts:
    download:
      warning: 25.0   # Adjust based on your internet plan
      critical: 10.0
    upload:
      warning: 5.0    # Adjust based on your needs
      critical: 1.0

network:
  speed_test:
    primary: "httpfile"  # Use "librespeed" if you have servers
    backends:
      httpfile:
        options:
          download_url: "https://proof.ovh.net/files/100Mb.dat"
```

### Environment Variables

Alternatively, configure via environment variables:

```bash
# Database
export IMON_DATABASE_PATH="./data/internet-monitor.db"
export IMON_DATABASE_RETENTION_DAYS=90

# Monitoring
export IMON_MONITOR_PEAK_HOURS_START="09:00"
export IMON_MONITOR_PEAK_HOURS_END="18:00"

# Alerts
export IMON_MONITOR_ALERTS_DOWNLOAD_WARNING=25.0
export IMON_MONITOR_ALERTS_DOWNLOAD_CRITICAL=10.0

# Network tests
export IMON_NETWORK_SPEED_TEST_PRIMARY="httpfile"
export IMON_NETWORK_PING_HOST="8.8.8.8"
```

## 🔧 System Service Setup

### Linux (systemd)

Create a dedicated user and service:

```bash
# Create system user
sudo useradd --system --home-dir /opt/internet-monitor --create-home --shell /bin/false internet-monitor

# Create directories
sudo mkdir -p /opt/internet-monitor/{data,logs,configs}
sudo chown -R internet-monitor:internet-monitor /opt/internet-monitor

# Copy binary and config
sudo cp /usr/local/bin/internet-monitor /opt/internet-monitor/
sudo cp config.yaml /opt/internet-monitor/

# Create systemd service
sudo tee /etc/systemd/system/internet-monitor.service > /dev/null <<EOF
[Unit]
Description=Internet Monitor
Documentation=https://github.com/cgorricho/internet-monitor
After=network.target
Wants=network.target

[Service]
Type=simple
User=internet-monitor
Group=internet-monitor
WorkingDirectory=/opt/internet-monitor
ExecStart=/opt/internet-monitor/internet-monitor monitor
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=internet-monitor

# Security settings
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/internet-monitor
PrivateTmp=yes
PrivateDevices=yes
ProtectControlGroups=yes
ProtectKernelModules=yes
ProtectKernelTunables=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes

[Install]
WantedBy=multi-user.target
EOF

# Initialize database as the service user
sudo -u internet-monitor /opt/internet-monitor/internet-monitor init

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable internet-monitor
sudo systemctl start internet-monitor

# Check status
sudo systemctl status internet-monitor
```

### Windows Service

```powershell
# Run PowerShell as Administrator

# Create service directory
New-Item -Path "C:\Program Files\Internet Monitor" -ItemType Directory -Force
Copy-Item internet-monitor.exe "C:\Program Files\Internet Monitor\"

# Install using sc command
sc create "Internet Monitor" binPath= "C:\Program Files\Internet Monitor\internet-monitor.exe monitor" start= auto

# Or use NSSM (Non-Sucking Service Manager)
# Download NSSM from https://nssm.cc/
nssm install "Internet Monitor" "C:\Program Files\Internet Monitor\internet-monitor.exe"
nssm set "Internet Monitor" Arguments "monitor"
nssm set "Internet Monitor" AppDirectory "C:\Program Files\Internet Monitor"
nssm set "Internet Monitor" DisplayName "Internet Connection Monitor"
nssm set "Internet Monitor" Description "Monitors internet connection performance"

# Start the service
net start "Internet Monitor"
```

### macOS (launchd)

```bash
# Create service user (optional)
sudo dscl . -create /Users/internet-monitor
sudo dscl . -create /Users/internet-monitor UserShell /usr/bin/false
sudo dscl . -create /Users/internet-monitor RealName "Internet Monitor Service"
sudo dscl . -create /Users/internet-monitor UniqueID 501
sudo dscl . -create /Users/internet-monitor PrimaryGroupID 20

# Create directories
sudo mkdir -p /opt/internet-monitor/{data,logs,configs}
sudo chown -R internet-monitor:staff /opt/internet-monitor

# Create launchd plist
sudo tee /Library/LaunchDaemons/com.cgorricho.internet-monitor.plist > /dev/null <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.cgorricho.internet-monitor</string>
    <key>ProgramArguments</key>
    <array>
        <string>/opt/internet-monitor/internet-monitor</string>
        <string>monitor</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/opt/internet-monitor</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>UserName</key>
    <string>internet-monitor</string>
    <key>StandardOutPath</key>
    <string>/opt/internet-monitor/logs/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/opt/internet-monitor/logs/stderr.log</string>
</dict>
</plist>
EOF

# Load and start service
sudo launchctl load /Library/LaunchDaemons/com.cgorricho.internet-monitor.plist
sudo launchctl start com.cgorricho.internet-monitor
```

## 🔗 Machine Pairing Setup

### Server Machine (e.g., Office)

```bash
# Enable pairing in configuration
echo "pairing:\n  enabled: true" >> config.yaml

# Start API server
internet-monitor serve &

# Generate pairing code
internet-monitor pair
# Output: Pairing code: ABC123 (expires in 5 minutes)
```

### Client Machine (e.g., Home)

```bash
# Join using pairing code
internet-monitor pair ABC123

# Verify pairing
internet-monitor status
# Should show paired machines
```

### Generate Comparative Dashboard

```bash
# On either machine
./scripts/dashboard.sh --compare
```

## 📊 Dashboard Generation

### Basic Usage

```bash
# Generate 24-hour dashboard
./scripts/dashboard.sh

# Generate weekly report
./scripts/dashboard.sh --hours 168 --output weekly-report.html

# Generate without opening browser
./scripts/dashboard.sh --no-browser
```

### Automated Reports

Create cron jobs for automated reporting:

```bash
# Edit crontab
crontab -e

# Add daily report generation (8 AM every day)
0 8 * * * /path/to/internet-monitor/scripts/dashboard.sh --output daily-$(date +\%Y-\%m-\%d).html --no-browser

# Add weekly report (Monday 9 AM)
0 9 * * 1 /path/to/internet-monitor/scripts/dashboard.sh --hours 168 --output weekly-$(date +\%Y-W\%V).html --no-browser
```

## 🔍 Monitoring and Logs

### Log Locations

- **Linux**: `/opt/internet-monitor/logs/internet-monitor.log`
- **Windows**: `C:\Program Files\Internet Monitor\logs\`
- **macOS**: `/opt/internet-monitor/logs/`

### Log Analysis

```bash
# Check recent activity
tail -f /opt/internet-monitor/logs/internet-monitor.log

# Find errors
grep "ERROR" /opt/internet-monitor/logs/internet-monitor.log

# Monitor real-time activity
journalctl -u internet-monitor -f  # Linux systemd
```

### Health Monitoring

```bash
# Check service status
systemctl status internet-monitor  # Linux
sc query "Internet Monitor"        # Windows
launchctl list | grep internet     # macOS

# Check database statistics
internet-monitor status
```

## 🚨 Troubleshooting

### Common Installation Issues

**Binary not found:**
```bash
# Verify PATH
echo $PATH
which internet-monitor

# Check binary permissions
ls -la /usr/local/bin/internet-monitor
```

**Permission denied:**
```bash
# Fix permissions
chmod +x internet-monitor
sudo chown root:root /usr/local/bin/internet-monitor
```

**Database initialization fails:**
```bash
# Check directory permissions
mkdir -p data logs
chmod 755 data logs

# Verify SQLite works
sqlite3 test.db "CREATE TABLE test (id INTEGER);"
rm test.db
```

### Network Test Issues

**Speed test failures:**
```bash
# Test connectivity
curl -I https://proof.ovh.net/files/100Mb.dat

# Check network tools
ping 8.8.8.8
nslookup google.com
iwconfig  # WiFi signal
```

**WiFi signal not detected:**
```bash
# Install wireless tools
sudo apt install wireless-tools  # Ubuntu/Debian
sudo yum install wireless-tools   # CentOS/RHEL

# Check wireless interfaces
iwconfig
ip link show
```

### Service Issues

**Service won't start:**
```bash
# Check service status and logs
sudo systemctl status internet-monitor
sudo journalctl -u internet-monitor -f

# Manual start for debugging
sudo -u internet-monitor /opt/internet-monitor/internet-monitor monitor
```

**High resource usage:**
```bash
# Check configuration intervals
grep -A 5 "intervals:" config.yaml

# Monitor resource usage
top -p $(pgrep internet-monitor)
```

## 🔒 Security Considerations

### File Permissions

```bash
# Secure configuration files
chmod 600 config.yaml
chmod 700 data/ logs/

# Service user permissions
sudo chown -R internet-monitor:internet-monitor /opt/internet-monitor
sudo chmod -R 755 /opt/internet-monitor
sudo chmod 600 /opt/internet-monitor/config.yaml
```

### Network Security

```bash
# Firewall for pairing (if needed)
sudo ufw allow 8080/tcp  # Ubuntu/Debian
sudo firewall-cmd --permanent --add-port=8080/tcp  # CentOS/RHEL
```

### Certificate Management

For secure pairing with TLS:

```bash
# Generate self-signed certificates
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

# Configure TLS in config.yaml
server:
  tls_cert: "./certs/server.crt"
  tls_key: "./certs/server.key"
```

## 📞 Support

If you encounter issues:

1. **Check Logs**: Review application logs for error messages
2. **GitHub Issues**: Report bugs at https://github.com/cgorricho/internet-monitor/issues
3. **Documentation**: Refer to README.md and other docs/
4. **Community**: Check existing issues for similar problems

---

**Installation Complete!** 🎉

Your Internet Monitor is now ready to track network performance and generate beautiful reports.

Next steps:
- Generate your first dashboard: `./scripts/dashboard.sh`
- Set up machine pairing if needed
- Schedule automated reports
- Monitor the logs for any issues