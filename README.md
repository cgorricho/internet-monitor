# Internet Connection Monitor

Real-time internet performance monitoring with an interactive Plotly Dash dashboard. Tracks download speed, upload speed, latency, jitter, packet loss, WiFi signal strength, and DNS resolution — with intelligent peak-hour scheduling and 3-month data retention.

## Why Build This?

ISP-provided "speed tests" give you a single point-in-time measurement. They don't tell you that your connection drops every Tuesday at 2pm, that your WiFi signal degrades when the microwave runs, or that your DNS resolution time spikes during peak hours.

This monitor runs continuously, captures the patterns, and surfaces the insights.

## What It Monitors

### Performance Metrics
| Metric | What It Reveals |
|--------|----------------|
| **Download Speed** (Mbps) | Bandwidth available for streaming, downloads |
| **Upload Speed** (Mbps) | Critical for video calls and remote work |
| **Ping/Latency** (ms) | Network responsiveness — affects real-time applications |
| **Jitter** (ms) | Latency stability — high jitter = choppy video calls |
| **Packet Loss** (%) | Connection reliability — even 1% causes noticeable issues |

### Infrastructure Health
| Metric | What It Reveals |
|--------|----------------|
| **WiFi Signal** (dBm) | Local wireless quality — is the problem your ISP or your router? |
| **DNS Resolution** (ms) | Web browsing performance — slow DNS = slow page loads |
| **Public IP** | ISP connection stability — IP changes indicate reconnections |
| **Uptime** (%) | Overall reliability trend |

## Smart Scheduling

- **Peak hours** (configurable): Higher frequency monitoring during work hours
- **Off-peak**: Reduced frequency to minimize bandwidth impact
- **Automatic adjustment**: Adapts based on time of day

## Dashboard Features

- **Time-series visualizations**: All metrics over configurable time ranges
- **Peak vs. off-peak comparison**: Side-by-side performance analysis
- **Alert system**: Configurable thresholds for degradation detection
- **3-month retention**: Automatic cleanup of old data
- **Real-time updates**: Dashboard refreshes with latest measurements

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Monitoring | Python (speedtest-cli, ping, WiFi tools) |
| Dashboard | Plotly Dash |
| Storage | Local SQLite / CSV |
| Scheduling | Built-in with peak-hour awareness |

## Quick Start

```bash
pip install -r requirements.txt
python monitor.py      # Start monitoring (background)
python dashboard.py    # Launch dashboard
```

Open `http://localhost:8050` to view the dashboard.
