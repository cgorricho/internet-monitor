# 🏗️ Architecture Overview - Internet Connection Monitor

**Last Updated**: October 10, 2024  
**Version**: 1.0  
**Maintainer**: Carlos Gorricho

## 🎯 System Overview

The Internet Connection Monitor is a modular Python application designed to continuously monitor network performance with intelligent scheduling and real-time visualization capabilities.

## 🧩 Component Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Dashboard     │    │    Monitor      │    │   Database      │
│   (dash.py)     │    │  (monitor.py)   │    │ (database.py)   │
│                 │    │                 │    │                 │
│ • Web Interface │◄───┤ • Scheduler     │───►│ • SQLite Store  │
│ • Time-series   │    │ • Data Collection│    │ • Data Retention│
│ • Alert Display │    │ • Threshold     │    │ • Query Helpers │
└─────────────────┘    │   Evaluation    │    └─────────────────┘
         ▲              └─────────────────┘             ▲
         │                       │                      │
         │                       ▼                      │
         │              ┌─────────────────┐             │
         │              │  Network Tests  │             │
         └──────────────┤ (network_tests) │─────────────┘
                        │                 │
                        │ • Speed Tests   │
                        │ • Ping/Latency  │
                        │ • DNS Resolution│
                        │ • WiFi Signal   │
                        └─────────────────┘
```

## 🔧 Core Components

### 1. Configuration Management (`config.py`)
**Purpose**: Centralized configuration and environment variable management  
**Responsibilities**:
- Load environment variables from `.env` file
- Define default values and validation
- Provide configuration constants to all modules
- Manage alert thresholds and monitoring schedules

```python
# Key Configuration Areas
- Database paths and settings
- Network test parameters  
- Alert thresholds (download, upload, ping, etc.)
- Peak hour definitions
- Dashboard settings
```

### 2. Database Layer (`database.py`)
**Purpose**: Data persistence and retrieval operations  
**Technology**: SQLite with sqlite-utils wrapper  
**Schema Design**:

```sql
-- measurements table
CREATE TABLE measurements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    download_mbps REAL,
    upload_mbps REAL,
    ping_ms REAL,
    jitter_ms REAL,
    packet_loss_percent REAL,
    dns_resolution_ms REAL,
    wifi_signal_dbm REAL,
    public_ip TEXT,
    alert_level TEXT,
    INDEX idx_timestamp (timestamp)
);
```

**Key Functions**:
- `init_db()`: Database initialization and schema creation
- `insert_measurement(data)`: Store network measurements
- `fetch_measurements(time_range)`: Retrieve historical data
- `purge_old_data(days)`: Automatic cleanup of old records

### 3. Network Testing Module (`network_tests.py`)
**Purpose**: Execute network performance measurements  
**Dependencies**: Ookla Speedtest CLI, system network utilities

**Core Functions**:
```python
def run_speedtest() -> dict:
    """Full bandwidth test with jitter and packet loss"""
    
def quick_ping(host='8.8.8.8') -> float:
    """Fast latency check for frequent monitoring"""
    
def get_dns_resolution_time(domain='google.com') -> float:
    """DNS performance measurement"""
    
def get_wifi_signal_strength() -> float:
    """Local WiFi connection quality"""
    
def get_public_ip() -> str:
    """ISP-assigned IP address tracking"""
```

### 4. Background Monitoring (`monitor.py`)
**Purpose**: Scheduled data collection with intelligent frequency  
**Technology**: APScheduler (Advanced Python Scheduler)

**Scheduling Logic**:
```python
# Peak Hours (9 AM - 6 PM)
- Quick tests: Every 2 minutes
- Full speed tests: Every 10 minutes
- Analysis/cleanup: Every 30 minutes

# Off-Peak Hours
- Quick tests: Every 5 minutes  
- Full speed tests: Every 20 minutes
- Analysis/cleanup: Every hour
```

**Job Types**:
- **Quick Monitoring**: Ping, DNS, WiFi signal (lightweight)
- **Comprehensive Testing**: Full speed test with all metrics
- **Data Analysis**: Threshold evaluation, alert generation, cleanup

### 5. Dashboard Interface (`dashboard.py`)
**Purpose**: Real-time web-based visualization and monitoring  
**Technology**: Plotly Dash with Bootstrap components

**Layout Structure**:
```
┌─────────────────────────────────────────────┐
│                Header/Navigation            │
├─────────────────┬───────────────────────────┤
│   Current       │                           │
│   Status        │      Time-Series          │
│   Panel         │      Visualizations       │
│                 │                           │
├─────────────────┼───────────────────────────┤
│   Alert         │      Controls             │
│   Messages      │      (Time Range,         │
│                 │       Parameters)         │
└─────────────────┴───────────────────────────┘
```

**Key Features**:
- Real-time data updates (60-second intervals)
- Interactive time-series charts for all metrics
- Peak vs off-peak performance comparison tabs
- Alert panel with color-coded issue notifications
- Mobile-responsive design

## 📊 Data Flow

### Monitoring Flow
```
1. Scheduler triggers network test
2. network_tests.py executes measurements
3. Results evaluated against thresholds
4. Data stored in SQLite database
5. Alerts generated if thresholds exceeded
6. Log entries created for audit trail
```

### Dashboard Flow
```
1. User accesses web interface
2. Dashboard queries recent data from database
3. Data processed and formatted for visualization  
4. Charts and metrics rendered in browser
5. Auto-refresh updates every 60 seconds
```

## 🚀 Deployment Architecture

### WSL Development Environment
```
WSL Ubuntu 24.04
├── Python 3.13.5 Virtual Environment
├── SQLite Database (local file storage)
├── Background Monitor Service
│   └── Runs as Python process
└── Dashboard Web Service  
    └── Accessible at localhost:8050
```

### Process Management
- **Monitor Process**: Long-running background scheduler
- **Dashboard Process**: Flask development server
- **System Integration**: Optional systemd service files

## 🔒 Security Considerations

### Data Protection
- **Local Storage**: All data stored locally in SQLite
- **No External Dependencies**: No cloud services or external APIs
- **Network Isolation**: Only outbound connections for testing

### Configuration Security
- **Environment Variables**: Sensitive settings in `.env` file
- **File Permissions**: Restricted access to configuration and logs
- **No Hardcoded Secrets**: All credentials externalized

## 📈 Performance Characteristics

### Resource Usage
- **CPU**: Minimal baseline usage, spikes during speed tests
- **Memory**: ~50-100MB typical usage
- **Disk**: ~1MB per day of measurements
- **Network**: Controlled test traffic based on schedule

### Scalability Considerations
- **Single Machine**: Designed for local monitoring only
- **Data Retention**: 90-day automatic cleanup prevents unbounded growth
- **Database Performance**: Indexed queries for historical data retrieval

## 🔧 Technology Stack

### Core Technologies
- **Language**: Python 3.13.5
- **Web Framework**: Plotly Dash + Flask
- **Database**: SQLite with sqlite-utils
- **Scheduling**: APScheduler
- **Visualization**: Plotly.js

### System Dependencies
- **Network Testing**: Ookla Speedtest CLI
- **Network Utilities**: ping, dig, iwconfig
- **System Tools**: Standard Linux networking stack

## 🎯 Design Principles

### Modularity
- **Separation of Concerns**: Each module has single responsibility
- **Loose Coupling**: Components interact through well-defined interfaces  
- **Configuration-Driven**: Behavior controlled through environment variables

### Reliability
- **Error Handling**: Graceful degradation when network tests fail
- **Data Integrity**: Transactional database operations
- **Recovery**: Automatic restart capabilities for monitoring

### Maintainability
- **Clear Structure**: Logical organization of code and configuration
- **Documentation**: Comprehensive inline and external documentation
- **Testing**: Unit and integration test coverage

## 🔮 Future Architecture Evolution

### Phase 2: Enhanced Analytics
- **Data Processing Pipeline**: Statistical analysis and trend detection
- **Machine Learning**: Anomaly detection and predictive capabilities
- **Advanced Visualization**: Custom charts and analysis tools

### Phase 3: Multi-Instance Support  
- **Configuration Management**: Support for multiple monitoring locations
- **Data Aggregation**: Centralized collection and analysis
- **Distributed Dashboard**: Multi-location performance comparison

---

**Architecture Review Cycle**: Quarterly  
**Next Review**: January 2025  
**Change Management**: All architectural changes documented in this file