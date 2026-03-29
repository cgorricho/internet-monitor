package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection and operations
type Database struct {
	db *sqlx.DB
}

// Measurement represents a network measurement record
type Measurement struct {
	ID              int64     `db:"id" json:"id"`
	MachineID       string    `db:"machine_id" json:"machine_id"`
	Timestamp       time.Time `db:"timestamp" json:"timestamp"`
	TestType        string    `db:"test_type" json:"test_type"`
	DownloadMbps    *float64  `db:"download_mbps" json:"download_mbps"`
	UploadMbps      *float64  `db:"upload_mbps" json:"upload_mbps"`
	PingMs          *float64  `db:"ping_ms" json:"ping_ms"`
	JitterMs        *float64  `db:"jitter_ms" json:"jitter_ms"`
	PacketLossPct   *float64  `db:"packet_loss_pct" json:"packet_loss_pct"`
	DNSResolutionMs *float64  `db:"dns_resolution_ms" json:"dns_resolution_ms"`
	WiFiSignalDbm   *float64  `db:"wifi_signal_dbm" json:"wifi_signal_dbm"`
	PublicIP        *string   `db:"public_ip" json:"public_ip"`
	AlertLevel      string    `db:"alert_level" json:"alert_level"`
	ServerName      *string   `db:"server_name" json:"server_name"`
	ServerLocation  *string   `db:"server_location" json:"server_location"`
	Success         bool      `db:"success" json:"success"`
	ErrorMessage    *string   `db:"error_message" json:"error_message"`
}

// Alert represents an alert record
type Alert struct {
	ID            int64     `db:"id" json:"id"`
	MeasurementID *int64    `db:"measurement_id" json:"measurement_id"`
	AlertType     string    `db:"alert_type" json:"alert_type"`
	AlertLevel    string    `db:"alert_level" json:"alert_level"`
	MetricName    string    `db:"metric_name" json:"metric_name"`
	MetricValue   float64   `db:"metric_value" json:"metric_value"`
	ThresholdValue float64  `db:"threshold_value" json:"threshold_value"`
	Message       string    `db:"message" json:"message"`
	Resolved      bool      `db:"resolved" json:"resolved"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	ResolvedAt    *time.Time `db:"resolved_at" json:"resolved_at"`
}

// PairPeer represents a paired machine
type PairPeer struct {
	ID          int64     `db:"id" json:"id"`
	MachineID   string    `db:"machine_id" json:"machine_id"`
	DisplayName string    `db:"display_name" json:"display_name"`
	Hostname    string    `db:"hostname" json:"hostname"`
	APIEndpoint string    `db:"api_endpoint" json:"api_endpoint"`
	PublicKey   string    `db:"public_key" json:"public_key"`
	LastSeen    time.Time `db:"last_seen" json:"last_seen"`
	Active      bool      `db:"active" json:"active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// Stats represents database statistics
type Stats struct {
	MeasurementCount int64        `json:"measurement_count"`
	AlertCount       int64        `json:"alert_count"`
	PeerCount        int64        `json:"peer_count"`
	DatabaseSizeMB   float64      `json:"database_size_mb"`
	LastMeasurement  *Measurement `json:"last_measurement"`
	DataRetentionDays int         `json:"data_retention_days"`
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	db, err := sqlx.Connect("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS measurements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			machine_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			test_type TEXT NOT NULL DEFAULT 'full',
			download_mbps REAL,
			upload_mbps REAL,
			ping_ms REAL,
			jitter_ms REAL,
			packet_loss_pct REAL,
			dns_resolution_ms REAL,
			wifi_signal_dbm REAL,
			public_ip TEXT,
			alert_level TEXT NOT NULL DEFAULT 'normal',
			server_name TEXT,
			server_location TEXT,
			success BOOLEAN NOT NULL DEFAULT 1,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_measurements_timestamp ON measurements(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_measurements_machine_id ON measurements(machine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_measurements_alert_level ON measurements(alert_level)`,
		`CREATE INDEX IF NOT EXISTS idx_measurements_test_type ON measurements(test_type)`,
		
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			measurement_id INTEGER,
			alert_type TEXT NOT NULL,
			alert_level TEXT NOT NULL,
			metric_name TEXT NOT NULL,
			metric_value REAL NOT NULL,
			threshold_value REAL NOT NULL,
			message TEXT NOT NULL,
			resolved BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME,
			FOREIGN KEY (measurement_id) REFERENCES measurements (id)
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_level ON alerts(alert_level)`,
		
		`CREATE TABLE IF NOT EXISTS pair_peers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			machine_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			hostname TEXT NOT NULL,
			api_endpoint TEXT NOT NULL,
			public_key TEXT NOT NULL,
			last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_pair_peers_machine_id ON pair_peers(machine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pair_peers_active ON pair_peers(active)`,
	}

	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// InsertMeasurement inserts a new measurement record
func (d *Database) InsertMeasurement(m *Measurement) error {
	query := `INSERT INTO measurements (
		machine_id, timestamp, test_type, download_mbps, upload_mbps, ping_ms, jitter_ms,
		packet_loss_pct, dns_resolution_ms, wifi_signal_dbm, public_ip, alert_level,
		server_name, server_location, success, error_message
	) VALUES (
		:machine_id, :timestamp, :test_type, :download_mbps, :upload_mbps, :ping_ms, :jitter_ms,
		:packet_loss_pct, :dns_resolution_ms, :wifi_signal_dbm, :public_ip, :alert_level,
		:server_name, :server_location, :success, :error_message
	)`

	result, err := d.db.NamedExec(query, m)
	if err != nil {
		return fmt.Errorf("failed to insert measurement: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	m.ID = id
	return nil
}

// GetMeasurements retrieves measurements with optional filtering
func (d *Database) GetMeasurements(machineID string, hours int, testType string, limit int) ([]*Measurement, error) {
	var measurements []*Measurement
	
	query := "SELECT * FROM measurements WHERE 1=1"
	args := make(map[string]interface{})
	
	if machineID != "" {
		query += " AND machine_id = :machine_id"
		args["machine_id"] = machineID
	}
	
	if hours > 0 {
		cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
		query += " AND timestamp >= :cutoff"
		args["cutoff"] = cutoff
	}
	
	if testType != "" {
		query += " AND test_type = :test_type"
		args["test_type"] = testType
	}
	
	query += " ORDER BY timestamp DESC"
	
	if limit > 0 {
		query += " LIMIT :limit"
		args["limit"] = limit
	}

	rows, err := d.db.NamedQuery(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query measurements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Measurement
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan measurement: %w", err)
		}
		measurements = append(measurements, &m)
	}

	return measurements, nil
}

// GetLatestMeasurement retrieves the most recent measurement
func (d *Database) GetLatestMeasurement(machineID string) (*Measurement, error) {
	var m Measurement
	
	query := "SELECT * FROM measurements WHERE machine_id = $1 ORDER BY timestamp DESC LIMIT 1"
	err := d.db.Get(&m, query, machineID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest measurement: %w", err)
	}

	return &m, nil
}

// InsertAlert inserts a new alert record
func (d *Database) InsertAlert(alert *Alert) error {
	query := `INSERT INTO alerts (
		measurement_id, alert_type, alert_level, metric_name, metric_value,
		threshold_value, message, resolved, created_at
	) VALUES (
		:measurement_id, :alert_type, :alert_level, :metric_name, :metric_value,
		:threshold_value, :message, :resolved, :created_at
	)`

	result, err := d.db.NamedExec(query, alert)
	if err != nil {
		return fmt.Errorf("failed to insert alert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	alert.ID = id
	return nil
}

// GetAlerts retrieves alerts with optional filtering
func (d *Database) GetAlerts(hours int, resolved *bool, alertLevel string, limit int) ([]*Alert, error) {
	var alerts []*Alert
	
	query := "SELECT * FROM alerts WHERE 1=1"
	args := make(map[string]interface{})
	
	if hours > 0 {
		cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
		query += " AND created_at >= :cutoff"
		args["cutoff"] = cutoff
	}
	
	if resolved != nil {
		query += " AND resolved = :resolved"
		args["resolved"] = *resolved
	}
	
	if alertLevel != "" {
		query += " AND alert_level = :alert_level"
		args["alert_level"] = alertLevel
	}
	
	query += " ORDER BY created_at DESC"
	
	if limit > 0 {
		query += " LIMIT :limit"
		args["limit"] = limit
	}

	rows, err := d.db.NamedQuery(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a Alert
		if err := rows.StructScan(&a); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}

	return alerts, nil
}

// UpsertPeer inserts or updates a peer record
func (d *Database) UpsertPeer(peer *PairPeer) error {
	query := `INSERT INTO pair_peers (
		machine_id, display_name, hostname, api_endpoint, public_key, last_seen, active
	) VALUES (
		:machine_id, :display_name, :hostname, :api_endpoint, :public_key, :last_seen, :active
	) ON CONFLICT(machine_id) DO UPDATE SET
		display_name = excluded.display_name,
		hostname = excluded.hostname,
		api_endpoint = excluded.api_endpoint,
		public_key = excluded.public_key,
		last_seen = excluded.last_seen,
		active = excluded.active`

	result, err := d.db.NamedExec(query, peer)
	if err != nil {
		return fmt.Errorf("failed to upsert peer: %w", err)
	}

	if peer.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert ID: %w", err)
		}
		peer.ID = id
	}

	return nil
}

// GetPeers retrieves all active peers
func (d *Database) GetPeers() ([]*PairPeer, error) {
	var peers []*PairPeer
	
	err := d.db.Select(&peers, "SELECT * FROM pair_peers WHERE active = 1 ORDER BY display_name")
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}

	return peers, nil
}

// DeletePeer marks a peer as inactive
func (d *Database) DeletePeer(machineID string) error {
	_, err := d.db.Exec("UPDATE pair_peers SET active = 0 WHERE machine_id = ?", machineID)
	if err != nil {
		return fmt.Errorf("failed to delete peer: %w", err)
	}
	
	return nil
}

// CleanupOldData removes old measurements and alerts
func (d *Database) CleanupOldData(retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	
	// Delete old measurements
	result, err := d.db.Exec("DELETE FROM measurements WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old measurements: %w", err)
	}
	
	measurementsDeleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	// Delete old alerts
	_, err = d.db.Exec("DELETE FROM alerts WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old alerts: %w", err)
	}
	
	// Run VACUUM to reclaim space
	if _, err := d.db.Exec("VACUUM"); err != nil {
		return 0, fmt.Errorf("failed to vacuum database: %w", err)
	}
	
	return measurementsDeleted, nil
}

// GetStats returns database statistics
func (d *Database) GetStats() (*Stats, error) {
	stats := &Stats{}
	
	// Get measurement count
	err := d.db.Get(&stats.MeasurementCount, "SELECT COUNT(*) FROM measurements")
	if err != nil {
		return nil, fmt.Errorf("failed to get measurement count: %w", err)
	}
	
	// Get alert count
	err = d.db.Get(&stats.AlertCount, "SELECT COUNT(*) FROM alerts")
	if err != nil {
		return nil, fmt.Errorf("failed to get alert count: %w", err)
	}
	
	// Get peer count
	err = d.db.Get(&stats.PeerCount, "SELECT COUNT(*) FROM pair_peers WHERE active = 1")
	if err != nil {
		return nil, fmt.Errorf("failed to get peer count: %w", err)
	}
	
	// Get latest measurement
	var latest Measurement
	err = d.db.Get(&latest, "SELECT * FROM measurements ORDER BY timestamp DESC LIMIT 1")
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get latest measurement: %w", err)
		}
	} else {
		stats.LastMeasurement = &latest
	}
	
	return stats, nil
}

// GetDatabaseSize returns the database file size in MB
func (d *Database) GetDatabaseSize(dbPath string) (float64, error) {
	info, err := os.Stat(dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat database file: %w", err)
	}
	
	return float64(info.Size()) / (1024 * 1024), nil
}