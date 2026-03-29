package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Monitor  MonitorConfig  `mapstructure:"monitor"`
	Network  NetworkConfig  `mapstructure:"network"`
	Server   ServerConfig   `mapstructure:"server"`
	Pairing  PairingConfig  `mapstructure:"pairing"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path            string        `mapstructure:"path"`
	RetentionDays   int          `mapstructure:"retention_days"`
	BackupPath      string       `mapstructure:"backup_path"`
	MaxConnections  int          `mapstructure:"max_connections"`
}

// MonitorConfig holds monitoring configuration
type MonitorConfig struct {
	PeakHours      PeakHoursConfig   `mapstructure:"peak_hours"`
	Intervals      IntervalsConfig   `mapstructure:"intervals"`
	Alerts         AlertsConfig      `mapstructure:"alerts"`
	MachineID      string           `mapstructure:"machine_id"`
	Timezone       string           `mapstructure:"timezone"`
}

// PeakHoursConfig defines peak monitoring hours
type PeakHoursConfig struct {
	Start time.Time `mapstructure:"start"`
	End   time.Time `mapstructure:"end"`
}

// IntervalsConfig defines monitoring intervals
type IntervalsConfig struct {
	Peak    IntervalSet `mapstructure:"peak"`
	OffPeak IntervalSet `mapstructure:"off_peak"`
}

// IntervalSet defines test intervals
type IntervalSet struct {
	Quick     time.Duration `mapstructure:"quick"`
	SpeedTest time.Duration `mapstructure:"speed_test"`
	Analysis  time.Duration `mapstructure:"analysis"`
}

// AlertsConfig defines alert thresholds
type AlertsConfig struct {
	Download   ThresholdConfig `mapstructure:"download"`
	Upload     ThresholdConfig `mapstructure:"upload"`
	Ping       ThresholdConfig `mapstructure:"ping"`
	PacketLoss ThresholdConfig `mapstructure:"packet_loss"`
	WiFi       ThresholdConfig `mapstructure:"wifi"`
	DNS        ThresholdConfig `mapstructure:"dns"`
}

// ThresholdConfig defines warning and critical thresholds
type ThresholdConfig struct {
	Warning  float64 `mapstructure:"warning"`
	Critical float64 `mapstructure:"critical"`
}

// NetworkConfig holds network testing configuration
type NetworkConfig struct {
	SpeedTest SpeedTestConfig `mapstructure:"speed_test"`
	Ping      PingConfig      `mapstructure:"ping"`
	DNS       DNSConfig       `mapstructure:"dns"`
	Timeout   time.Duration   `mapstructure:"timeout"`
	Retries   int            `mapstructure:"retries"`
}

// SpeedTestConfig defines speed test backends
type SpeedTestConfig struct {
	Primary   string              `mapstructure:"primary"`
	Fallbacks []string            `mapstructure:"fallbacks"`
	Backends  map[string]BackendConfig `mapstructure:"backends"`
}

// BackendConfig defines backend-specific configuration
type BackendConfig struct {
	Servers []string          `mapstructure:"servers"`
	Options map[string]string `mapstructure:"options"`
}

// PingConfig defines ping test configuration
type PingConfig struct {
	Host    string        `mapstructure:"host"`
	Count   int          `mapstructure:"count"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// DNSConfig defines DNS resolution test configuration
type DNSConfig struct {
	Domain  string        `mapstructure:"domain"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// ServerConfig holds API server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int          `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	TLSCert      string       `mapstructure:"tls_cert"`
	TLSKey       string       `mapstructure:"tls_key"`
}

// PairingConfig holds pairing system configuration
type PairingConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	TokenTTL      time.Duration `mapstructure:"token_ttl"`
	PeersFile     string       `mapstructure:"peers_file"`
	CertPath      string       `mapstructure:"cert_path"`
	KeyPath       string       `mapstructure:"key_path"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// Load loads configuration from various sources
func Load() (*Config, error) {
	v := viper.New()
	
	// Set defaults
	setDefaults(v)
	
	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("$HOME/.internet-monitor")
	v.AddConfigPath("/etc/internet-monitor")

	// Enable environment variables
	v.SetEnvPrefix("IMON")
	v.AutomaticEnv()

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Ensure required directories exist
	if err := ensureDirectories(&cfg); err != nil {
		return nil, fmt.Errorf("error creating directories: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Database defaults
	v.SetDefault("database.path", "./data/internet-monitor.db")
	v.SetDefault("database.retention_days", 90)
	v.SetDefault("database.backup_path", "./data/backups")
	v.SetDefault("database.max_connections", 10)

	// Monitor defaults
	v.SetDefault("monitor.peak_hours.start", "09:00")
	v.SetDefault("monitor.peak_hours.end", "18:00")
	v.SetDefault("monitor.intervals.peak.quick", "2m")
	v.SetDefault("monitor.intervals.peak.speed_test", "10m")
	v.SetDefault("monitor.intervals.peak.analysis", "30m")
	v.SetDefault("monitor.intervals.off_peak.quick", "5m")
	v.SetDefault("monitor.intervals.off_peak.speed_test", "20m")
	v.SetDefault("monitor.intervals.off_peak.analysis", "60m")
	v.SetDefault("monitor.machine_id", generateMachineID())
	v.SetDefault("monitor.timezone", "Local")

	// Alert thresholds
	v.SetDefault("monitor.alerts.download.warning", 25.0)
	v.SetDefault("monitor.alerts.download.critical", 10.0)
	v.SetDefault("monitor.alerts.upload.warning", 5.0)
	v.SetDefault("monitor.alerts.upload.critical", 1.0)
	v.SetDefault("monitor.alerts.ping.warning", 50.0)
	v.SetDefault("monitor.alerts.ping.critical", 100.0)
	v.SetDefault("monitor.alerts.packet_loss.warning", 1.0)
	v.SetDefault("monitor.alerts.packet_loss.critical", 3.0)
	v.SetDefault("monitor.alerts.wifi.warning", -70.0)
	v.SetDefault("monitor.alerts.wifi.critical", -80.0)
	v.SetDefault("monitor.alerts.dns.warning", 100.0)
	v.SetDefault("monitor.alerts.dns.critical", 500.0)

	// Network defaults
	v.SetDefault("network.speed_test.primary", "librespeed")
	v.SetDefault("network.speed_test.fallbacks", []string{"httpfile"})
	v.SetDefault("network.speed_test.backends.librespeed.servers", []string{})
	v.SetDefault("network.speed_test.backends.httpfile.options.download_url", "https://proof.ovh.net/files/100Mb.dat")
	v.SetDefault("network.speed_test.backends.httpfile.options.upload_url", "https://httpbin.org/post")
	v.SetDefault("network.ping.host", "8.8.8.8")
	v.SetDefault("network.ping.count", 4)
	v.SetDefault("network.ping.timeout", "5s")
	v.SetDefault("network.dns.domain", "google.com")
	v.SetDefault("network.dns.timeout", "5s")
	v.SetDefault("network.timeout", "60s")
	v.SetDefault("network.retries", 3)

	// Server defaults
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")

	// Pairing defaults
	v.SetDefault("pairing.enabled", false)
	v.SetDefault("pairing.token_ttl", "5m")
	v.SetDefault("pairing.peers_file", "./data/peers.json")
	v.SetDefault("pairing.cert_path", "./data/certs")
	v.SetDefault("pairing.key_path", "./data/keys")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.file", "./logs/internet-monitor.log")
	v.SetDefault("logging.max_size", 100)
	v.SetDefault("logging.max_backups", 5)
	v.SetDefault("logging.max_age", 30)
}

// ensureDirectories creates required directories
func ensureDirectories(cfg *Config) error {
	dirs := []string{
		filepath.Dir(cfg.Database.Path),
		cfg.Database.BackupPath,
		filepath.Dir(cfg.Logging.File),
		cfg.Pairing.CertPath,
		cfg.Pairing.KeyPath,
		filepath.Dir(cfg.Pairing.PeersFile),
	}

	for _, dir := range dirs {
		if dir == "" || dir == "." {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateMachineID generates a unique machine identifier
func generateMachineID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	
	// In a real implementation, you might want to include hardware info
	// for a more stable ID across renames
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}

// IsPeakHours returns true if the current time is within peak hours
func (cfg *Config) IsPeakHours(t time.Time) bool {
	startHour := cfg.Monitor.PeakHours.Start.Hour()
	startMin := cfg.Monitor.PeakHours.Start.Minute()
	endHour := cfg.Monitor.PeakHours.End.Hour()
	endMin := cfg.Monitor.PeakHours.End.Minute()
	
	currentMinutes := t.Hour()*60 + t.Minute()
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin
	
	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

// GetIntervals returns appropriate intervals based on peak hours
func (cfg *Config) GetIntervals(isPeak bool) IntervalSet {
	if isPeak {
		return cfg.Monitor.Intervals.Peak
	}
	return cfg.Monitor.Intervals.OffPeak
}

// GetAlertLevel returns alert level for a metric value
func (cfg *Config) GetAlertLevel(metric string, value float64) string {
	var threshold ThresholdConfig
	var higherIsBad bool = true
	
	switch metric {
	case "download":
		threshold = cfg.Monitor.Alerts.Download
		higherIsBad = false
	case "upload":
		threshold = cfg.Monitor.Alerts.Upload
		higherIsBad = false
	case "ping":
		threshold = cfg.Monitor.Alerts.Ping
	case "packet_loss":
		threshold = cfg.Monitor.Alerts.PacketLoss
	case "wifi":
		threshold = cfg.Monitor.Alerts.WiFi
		higherIsBad = false
	case "dns":
		threshold = cfg.Monitor.Alerts.DNS
	default:
		return "normal"
	}
	
	if higherIsBad {
		if value >= threshold.Critical {
			return "critical"
		} else if value >= threshold.Warning {
			return "warning"
		}
	} else {
		if value <= threshold.Critical {
			return "critical"
		} else if value <= threshold.Warning {
			return "warning"
		}
	}
	
	return "normal"
}