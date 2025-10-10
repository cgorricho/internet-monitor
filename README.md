# 📡 Internet Connection Monitor

A comprehensive Python application that monitors internet connection performance and displays real-time metrics through an interactive Plotly Dash dashboard.

## 🎯 Features

- **Real-time monitoring** of internet connection parameters
- **Smart scheduling** with increased frequency during peak office hours
- **Interactive dashboard** with time-series visualizations
- **WiFi signal strength** monitoring for local connection quality
- **Alert system** with configurable thresholds
- **3-month data retention** with automatic cleanup
- **Peak vs Off-Peak** performance comparisons

## 📊 Monitored Parameters

### Core Performance Metrics
- **Download Speed** (Mbps) - Primary bandwidth indicator
- **Upload Speed** (Mbps) - Critical for remote work and video calls
- **Ping/Latency** (ms) - Network responsiveness
- **Jitter** (ms) - Latency stability
- **Packet Loss** (%) - Connection reliability

### Infrastructure Health
- **WiFi Signal Strength** (dBm) - Local wireless quality
- **DNS Resolution Time** (ms) - Web browsing performance
- **Public IP Address** - ISP connection stability tracking
- **Connection Uptime** (%) - Overall reliability

## 🚀 Quick Start

### 1. System Dependencies
```bash
# Install system-level dependencies
sudo ./scripts/install_deps.sh
```

### 2. Python Environment
```bash
# Create and activate virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Install Python dependencies
pip install -r requirements.txt
```

### 3. Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit configuration as needed
nano .env
```

### 4. Database Initialization
```bash
# Initialize SQLite database
python -c "from src.database import init_db; init_db()"
```

### 5. Run the Application
```bash
# Start monitoring (in background)
./scripts/run_monitor.sh

# Launch dashboard (separate terminal)
./scripts/run_dashboard.sh
```

## 📈 Dashboard Access

Open your web browser and navigate to:
- **Local Access**: http://localhost:8050
- **Network Access**: http://[your-ip]:8050

## ⚙️ Configuration

The application uses environment variables for configuration. Key settings include:

- `PEAK_HOURS_START`: Start of peak monitoring (default: 09:00)
- `PEAK_HOURS_END`: End of peak monitoring (default: 18:00)
- `ALERT_DOWNLOAD_WARNING`: Download speed warning threshold (Mbps)
- `ALERT_PING_CRITICAL`: Critical ping threshold (ms)

See `.env.example` for all available options.

## 🔄 Monitoring Schedule

### Peak Office Hours (9 AM - 6 PM)
- **Quick Tests**: Every 2 minutes
- **Speed Tests**: Every 10 minutes
- **Analysis**: Every 30 minutes

### Off-Peak Hours
- **Quick Tests**: Every 5 minutes
- **Speed Tests**: Every 20 minutes
- **Analysis**: Every hour

## 📋 Requirements

- **Python**: 3.9+
- **OS**: Linux (WSL compatible)
- **Network**: Active internet connection
- **Storage**: ~50MB for 3 months of data

## 🛠️ Development

This project follows the development workflow defined in `DEVELOPMENT_WORKFLOW_GUIDE.md`:

- Feature development in branches
- Comprehensive testing
- Documentation-first approach
- Automated deployment scripts

## 📚 Documentation

- [Architecture Overview](docs/architecture.md)
- [Deployment Guide](docs/deployment.md)
- [Project Status](PROJECT_STATUS.md)
- [Improvement Plan](IMPROVEMENT_PLAN.md)

## 🚨 Troubleshooting

### Common Issues

1. **Speedtest fails**: Ensure Ookla CLI is installed and licensed accepted
2. **WiFi detection fails**: Check if `iwconfig` is available and user has permissions
3. **Dashboard not accessible**: Verify port 8050 is not in use

### Logs
Check application logs in the `logs/` directory for detailed error information.

## 📄 License

This project is intended for personal use and monitoring of your own internet connection.

---

**Created by**: Carlos Gorricho  
**Last Updated**: October 2024