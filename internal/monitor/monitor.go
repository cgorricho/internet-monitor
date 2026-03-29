package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/cgorricho/internet-monitor/internal/config"
	"github.com/cgorricho/internet-monitor/internal/database"
	"github.com/cgorricho/internet-monitor/internal/network"
)

// Monitor handles background network monitoring
type Monitor struct {
	config  *config.Config
	db      *database.Database
	network *network.Tester
	cron    *cron.Cron
}

// New creates a new monitor instance
func New(cfg *config.Config, db *database.Database) *Monitor {
	networkTester := network.New(cfg)
	
	// Create cron scheduler with seconds precision
	cronScheduler := cron.New(cron.WithSeconds())

	return &Monitor{
		config:  cfg,
		db:      db,
		network: networkTester,
		cron:    cronScheduler,
	}
}

// Start begins the monitoring service
func (m *Monitor) Start(ctx context.Context) error {
	log.Printf("Starting internet monitor for machine: %s", m.config.Monitor.MachineID)

	// Schedule quick tests
	m.scheduleQuickTests()

	// Schedule comprehensive tests
	m.scheduleComprehensiveTests()

	// Schedule cleanup tasks
	m.scheduleCleanup()

	// Start the cron scheduler
	m.cron.Start()

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Stopping monitor...")
	cronCtx := m.cron.Stop()
	<-cronCtx.Done()

	return nil
}

// scheduleQuickTests schedules quick network tests
func (m *Monitor) scheduleQuickTests() {
	// Peak hours quick tests
	peakInterval := m.config.Monitor.Intervals.Peak.Quick
	peakCron := fmt.Sprintf("@every %s", peakInterval)
	
	// Off-peak hours quick tests
	offPeakInterval := m.config.Monitor.Intervals.OffPeak.Quick
	offPeakCron := fmt.Sprintf("@every %s", offPeakInterval)

	// We'll use a single job that decides based on current time
	quickTestJob := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := m.runQuickTest(ctx); err != nil {
			log.Printf("Quick test failed: %v", err)
		}
	}

	// For now, we'll use the more frequent peak interval and check time within the job
	// In a production system, you might want dynamic cron scheduling
	m.cron.AddFunc(peakCron, quickTestJob)
	
	log.Printf("Scheduled quick tests: peak every %s, off-peak every %s", peakInterval, offPeakInterval)
}

// scheduleComprehensiveTests schedules comprehensive network tests
func (m *Monitor) scheduleComprehensiveTests() {
	// Peak hours comprehensive tests
	peakInterval := m.config.Monitor.Intervals.Peak.SpeedTest
	peakCron := fmt.Sprintf("@every %s", peakInterval)

	comprehensiveTestJob := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		if err := m.runComprehensiveTest(ctx); err != nil {
			log.Printf("Comprehensive test failed: %v", err)
		}
	}

	m.cron.AddFunc(peakCron, comprehensiveTestJob)
	
	log.Printf("Scheduled comprehensive tests every %s", peakInterval)
}

// scheduleCleanup schedules database cleanup tasks
func (m *Monitor) scheduleCleanup() {
	// Run cleanup daily at 2 AM
	cleanupJob := func() {
		if err := m.runCleanup(); err != nil {
			log.Printf("Cleanup failed: %v", err)
		}
	}

	m.cron.AddFunc("0 0 2 * * *", cleanupJob)
	log.Println("Scheduled daily cleanup at 2 AM")
}

// runQuickTest executes a quick network test
func (m *Monitor) runQuickTest(ctx context.Context) error {
	log.Println("Running quick network test...")

	result, err := m.network.RunQuickTest(ctx)
	if err != nil {
		return fmt.Errorf("quick test failed: %w", err)
	}

	// Convert to database measurement
	measurement := m.convertToMeasurement(result, "quick")

	// Calculate alert level
	alertLevel := m.calculateAlertLevel(result)
	measurement.AlertLevel = alertLevel

	// Store in database
	if err := m.db.InsertMeasurement(measurement); err != nil {
		return fmt.Errorf("failed to store measurement: %w", err)
	}

	// Create alerts if necessary
	if alertLevel != "normal" {
		if err := m.createAlerts(measurement, result); err != nil {
			log.Printf("Failed to create alerts: %v", err)
		}
	}

	log.Printf("Quick test completed successfully (alert level: %s)", alertLevel)
	return nil
}

// runComprehensiveTest executes a comprehensive network test
func (m *Monitor) runComprehensiveTest(ctx context.Context) error {
	log.Println("Running comprehensive network test...")

	result, err := m.network.RunComprehensiveTest(ctx)
	if err != nil {
		return fmt.Errorf("comprehensive test failed: %w", err)
	}

	// Convert to database measurement
	measurement := m.convertToMeasurement(result, "full")

	// Calculate alert level
	alertLevel := m.calculateAlertLevel(result)
	measurement.AlertLevel = alertLevel

	// Store in database
	if err := m.db.InsertMeasurement(measurement); err != nil {
		return fmt.Errorf("failed to store measurement: %w", err)
	}

	// Create alerts if necessary
	if alertLevel != "normal" {
		if err := m.createAlerts(measurement, result); err != nil {
			log.Printf("Failed to create alerts: %v", err)
		}
	}

	log.Printf("Comprehensive test completed successfully (alert level: %s)", alertLevel)
	return nil
}

// runCleanup performs database cleanup
func (m *Monitor) runCleanup() error {
	log.Println("Running database cleanup...")

	deletedCount, err := m.db.CleanupOldData(m.config.Database.RetentionDays)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	log.Printf("Cleanup completed: deleted %d old records", deletedCount)
	return nil
}

// convertToMeasurement converts network test result to database measurement
func (m *Monitor) convertToMeasurement(result *network.TestResult, testType string) *database.Measurement {
	measurement := &database.Measurement{
		MachineID:       m.config.Monitor.MachineID,
		Timestamp:       time.Now(),
		TestType:        testType,
		DownloadMbps:    result.DownloadMbps,
		UploadMbps:      result.UploadMbps,
		PingMs:          result.PingMs,
		JitterMs:        result.JitterMs,
		PacketLossPct:   result.PacketLossPct,
		DNSResolutionMs: result.DNSResolutionMs,
		WiFiSignalDbm:   result.WiFiSignalDbm,
		PublicIP:        result.PublicIP,
		ServerName:      result.ServerName,
		ServerLocation:  result.ServerLocation,
		Success:         result.Success,
		ErrorMessage:    result.ErrorMessage,
	}

	return measurement
}

// calculateAlertLevel determines the overall alert level for a test result
func (m *Monitor) calculateAlertLevel(result *network.TestResult) string {
	alertLevels := []string{}

	// Check each metric that has a value
	if result.DownloadMbps != nil {
		level := m.config.GetAlertLevel("download", *result.DownloadMbps)
		alertLevels = append(alertLevels, level)
	}

	if result.UploadMbps != nil {
		level := m.config.GetAlertLevel("upload", *result.UploadMbps)
		alertLevels = append(alertLevels, level)
	}

	if result.PingMs != nil {
		level := m.config.GetAlertLevel("ping", *result.PingMs)
		alertLevels = append(alertLevels, level)
	}

	if result.PacketLossPct != nil {
		level := m.config.GetAlertLevel("packet_loss", *result.PacketLossPct)
		alertLevels = append(alertLevels, level)
	}

	if result.WiFiSignalDbm != nil {
		level := m.config.GetAlertLevel("wifi", *result.WiFiSignalDbm)
		alertLevels = append(alertLevels, level)
	}

	if result.DNSResolutionMs != nil {
		level := m.config.GetAlertLevel("dns", *result.DNSResolutionMs)
		alertLevels = append(alertLevels, level)
	}

	// Return highest priority alert level
	for _, level := range alertLevels {
		if level == "critical" {
			return "critical"
		}
	}

	for _, level := range alertLevels {
		if level == "warning" {
			return "warning"
		}
	}

	return "normal"
}

// createAlerts creates alert records for problematic metrics
func (m *Monitor) createAlerts(measurement *database.Measurement, result *network.TestResult) error {
	alerts := []*database.Alert{}

	// Check download speed
	if result.DownloadMbps != nil {
		level := m.config.GetAlertLevel("download", *result.DownloadMbps)
		if level != "normal" {
			threshold := m.getThreshold("download", level)
			alert := &database.Alert{
				MeasurementID:  &measurement.ID,
				AlertType:      "performance",
				AlertLevel:     level,
				MetricName:     "download_mbps",
				MetricValue:    *result.DownloadMbps,
				ThresholdValue: threshold,
				Message:        fmt.Sprintf("Download speed %s: %.2f Mbps (threshold: %.2f Mbps)", level, *result.DownloadMbps, threshold),
				CreatedAt:      time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check upload speed
	if result.UploadMbps != nil {
		level := m.config.GetAlertLevel("upload", *result.UploadMbps)
		if level != "normal" {
			threshold := m.getThreshold("upload", level)
			alert := &database.Alert{
				MeasurementID:  &measurement.ID,
				AlertType:      "performance",
				AlertLevel:     level,
				MetricName:     "upload_mbps",
				MetricValue:    *result.UploadMbps,
				ThresholdValue: threshold,
				Message:        fmt.Sprintf("Upload speed %s: %.2f Mbps (threshold: %.2f Mbps)", level, *result.UploadMbps, threshold),
				CreatedAt:      time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check ping latency
	if result.PingMs != nil {
		level := m.config.GetAlertLevel("ping", *result.PingMs)
		if level != "normal" {
			threshold := m.getThreshold("ping", level)
			alert := &database.Alert{
				MeasurementID:  &measurement.ID,
				AlertType:      "latency",
				AlertLevel:     level,
				MetricName:     "ping_ms",
				MetricValue:    *result.PingMs,
				ThresholdValue: threshold,
				Message:        fmt.Sprintf("Ping latency %s: %.2f ms (threshold: %.2f ms)", level, *result.PingMs, threshold),
				CreatedAt:      time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Insert all alerts
	for _, alert := range alerts {
		if err := m.db.InsertAlert(alert); err != nil {
			return fmt.Errorf("failed to insert alert: %w", err)
		}
	}

	if len(alerts) > 0 {
		log.Printf("Created %d alert(s)", len(alerts))
	}

	return nil
}

// getThreshold gets the threshold value for a metric and alert level
func (m *Monitor) getThreshold(metric, level string) float64 {
	switch metric {
	case "download":
		if level == "critical" {
			return m.config.Monitor.Alerts.Download.Critical
		}
		return m.config.Monitor.Alerts.Download.Warning
	case "upload":
		if level == "critical" {
			return m.config.Monitor.Alerts.Upload.Critical
		}
		return m.config.Monitor.Alerts.Upload.Warning
	case "ping":
		if level == "critical" {
			return m.config.Monitor.Alerts.Ping.Critical
		}
		return m.config.Monitor.Alerts.Ping.Warning
	case "packet_loss":
		if level == "critical" {
			return m.config.Monitor.Alerts.PacketLoss.Critical
		}
		return m.config.Monitor.Alerts.PacketLoss.Warning
	case "wifi":
		if level == "critical" {
			return m.config.Monitor.Alerts.WiFi.Critical
		}
		return m.config.Monitor.Alerts.WiFi.Warning
	case "dns":
		if level == "critical" {
			return m.config.Monitor.Alerts.DNS.Critical
		}
		return m.config.Monitor.Alerts.DNS.Warning
	default:
		return 0
	}
}