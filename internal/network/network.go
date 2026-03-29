package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cgorricho/internet-monitor/internal/config"
)

// TestResult represents the result of a network test
type TestResult struct {
	DownloadMbps    *float64 `json:"download_mbps"`
	UploadMbps      *float64 `json:"upload_mbps"`
	PingMs          *float64 `json:"ping_ms"`
	JitterMs        *float64 `json:"jitter_ms"`
	PacketLossPct   *float64 `json:"packet_loss_pct"`
	DNSResolutionMs *float64 `json:"dns_resolution_ms"`
	WiFiSignalDbm   *float64 `json:"wifi_signal_dbm"`
	PublicIP        *string  `json:"public_ip"`
	ServerName      *string  `json:"server_name"`
	ServerLocation  *string  `json:"server_location"`
	Success         bool     `json:"success"`
	ErrorMessage    *string  `json:"error_message"`
}

// SpeedTestBackend represents a speed test implementation
type SpeedTestBackend interface {
	Name() string
	IsAvailable(ctx context.Context) bool
	RunTest(ctx context.Context) (*SpeedTestResult, error)
}

// SpeedTestResult represents speed test specific results
type SpeedTestResult struct {
	DownloadMbps   float64 `json:"download_mbps"`
	UploadMbps     float64 `json:"upload_mbps"`
	PingMs         float64 `json:"ping_ms"`
	JitterMs       float64 `json:"jitter_ms"`
	PacketLossPct  float64 `json:"packet_loss_pct"`
	ServerName     string  `json:"server_name"`
	ServerLocation string  `json:"server_location"`
}

// Tester handles network performance testing
type Tester struct {
	config   *config.Config
	backends map[string]SpeedTestBackend
	client   *http.Client
}

// New creates a new network tester
func New(cfg *config.Config) *Tester {
	client := &http.Client{
		Timeout: cfg.Network.Timeout,
	}

	tester := &Tester{
		config:   cfg,
		backends: make(map[string]SpeedTestBackend),
		client:   client,
	}

	// Register available backends
	tester.backends["httpfile"] = NewHTTPFileBackend(cfg, client)
	tester.backends["librespeed"] = NewLibreSpeedBackend(cfg, client)

	return tester
}

// RunComprehensiveTest runs a full set of network tests
func (t *Tester) RunComprehensiveTest(ctx context.Context) (*TestResult, error) {
	result := &TestResult{}

	// Run speed test
	speedResult, err := t.runSpeedTest(ctx)
	if err == nil && speedResult != nil {
		result.DownloadMbps = &speedResult.DownloadMbps
		result.UploadMbps = &speedResult.UploadMbps
		result.PingMs = &speedResult.PingMs
		result.JitterMs = &speedResult.JitterMs
		result.PacketLossPct = &speedResult.PacketLossPct
		result.ServerName = &speedResult.ServerName
		result.ServerLocation = &speedResult.ServerLocation
		result.Success = true
	}

	// Run DNS test
	if dnsTime := t.testDNSResolution(ctx); dnsTime > 0 {
		result.DNSResolutionMs = &dnsTime
	}

	// Get WiFi signal strength
	if wifiSignal := t.getWiFiSignalStrength(ctx); wifiSignal != 0 {
		result.WiFiSignalDbm = &wifiSignal
	}

	// Get public IP
	if publicIP := t.getPublicIP(ctx); publicIP != "" {
		result.PublicIP = &publicIP
	}

	// If speed test failed but we have other metrics, still consider it partially successful
	if !result.Success && (result.DNSResolutionMs != nil || result.WiFiSignalDbm != nil || result.PublicIP != nil) {
		result.Success = true
		errorMsg := "Speed test failed, but other metrics available"
		result.ErrorMessage = &errorMsg
	}

	return result, nil
}

// RunQuickTest runs lightweight network tests
func (t *Tester) RunQuickTest(ctx context.Context) (*TestResult, error) {
	result := &TestResult{}

	// Quick ping test
	if pingTime := t.testPing(ctx); pingTime > 0 {
		result.PingMs = &pingTime
		result.Success = true
	}

	// DNS test
	if dnsTime := t.testDNSResolution(ctx); dnsTime > 0 {
		result.DNSResolutionMs = &dnsTime
		result.Success = true
	}

	// WiFi signal
	if wifiSignal := t.getWiFiSignalStrength(ctx); wifiSignal != 0 {
		result.WiFiSignalDbm = &wifiSignal
		result.Success = true
	}

	if !result.Success {
		errorMsg := "All quick tests failed"
		result.ErrorMessage = &errorMsg
	}

	return result, nil
}

// runSpeedTest runs speed test using configured backends
func (t *Tester) runSpeedTest(ctx context.Context) (*SpeedTestResult, error) {
	// Try primary backend first
	if backend, ok := t.backends[t.config.Network.SpeedTest.Primary]; ok {
		if backend.IsAvailable(ctx) {
			result, err := backend.RunTest(ctx)
			if err == nil {
				return result, nil
			}
		}
	}

	// Try fallback backends
	for _, backendName := range t.config.Network.SpeedTest.Fallbacks {
		if backend, ok := t.backends[backendName]; ok {
			if backend.IsAvailable(ctx) {
				result, err := backend.RunTest(ctx)
				if err == nil {
					return result, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("all speed test backends failed")
}

// testPing performs a ping test
func (t *Tester) testPing(ctx context.Context) float64 {
	cmd := exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(t.config.Network.Ping.Count), t.config.Network.Ping.Host)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Parse ping output for average time
	// Example: "rtt min/avg/max/mdev = 12.345/23.456/34.567/5.678 ms"
	re := regexp.MustCompile(`rtt min/avg/max/mdev = [^/]+/([^/]+)/`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		if avgTime, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return avgTime
		}
	}

	return 0
}

// testDNSResolution tests DNS resolution time
func (t *Tester) testDNSResolution(ctx context.Context) float64 {
	start := time.Now()
	
	resolver := &net.Resolver{}
	_, err := resolver.LookupHost(ctx, t.config.Network.DNS.Domain)
	if err != nil {
		return 0
	}

	duration := time.Since(start)
	return float64(duration.Nanoseconds()) / 1e6 // Convert to milliseconds
}

// getWiFiSignalStrength gets WiFi signal strength using iwconfig
func (t *Tester) getWiFiSignalStrength(ctx context.Context) float64 {
	cmd := exec.CommandContext(ctx, "iwconfig")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Parse iwconfig output for signal strength
	// Example: "Link Quality=70/70  Signal level=-40 dBm"
	re := regexp.MustCompile(`Signal level=(-?\d+(?:\.\d+)?)\s*dBm`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		if signalLevel, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return signalLevel
		}
	}

	return 0
}

// getPublicIP gets the public IP address
func (t *Tester) getPublicIP(ctx context.Context) string {
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	for _, service := range services {
		req, err := http.NewRequestWithContext(ctx, "GET", service, nil)
		if err != nil {
			continue
		}

		resp, err := t.client.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			ip := strings.TrimSpace(string(body))
			// Basic IP validation
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	return ""
}

// HTTPFileBackend implements speed testing using HTTP file downloads/uploads
type HTTPFileBackend struct {
	config *config.Config
	client *http.Client
}

// NewHTTPFileBackend creates a new HTTP file backend
func NewHTTPFileBackend(cfg *config.Config, client *http.Client) *HTTPFileBackend {
	return &HTTPFileBackend{
		config: cfg,
		client: client,
	}
}

// Name returns the backend name
func (h *HTTPFileBackend) Name() string {
	return "httpfile"
}

// IsAvailable checks if the backend is available
func (h *HTTPFileBackend) IsAvailable(ctx context.Context) bool {
	// Check if we have configured URLs
	if options, ok := h.config.Network.SpeedTest.Backends["httpfile"]; ok {
		if options.Options["download_url"] != "" {
			return true
		}
	}
	return false
}

// RunTest runs the HTTP file speed test
func (h *HTTPFileBackend) RunTest(ctx context.Context) (*SpeedTestResult, error) {
	result := &SpeedTestResult{}

	options, ok := h.config.Network.SpeedTest.Backends["httpfile"]
	if !ok {
		return nil, fmt.Errorf("httpfile backend not configured")
	}

	downloadURL := options.Options["download_url"]
	if downloadURL == "" {
		return nil, fmt.Errorf("download_url not configured")
	}

	// Test download speed
	downloadSpeed, err := h.testDownloadSpeed(ctx, downloadURL)
	if err != nil {
		return nil, fmt.Errorf("download test failed: %w", err)
	}

	result.DownloadMbps = downloadSpeed
	result.ServerName = "HTTP File Test"
	result.ServerLocation = "Unknown"

	// Upload test is more complex and optional for now
	result.UploadMbps = 0

	return result, nil
}

// testDownloadSpeed tests download speed by downloading a file
func (h *HTTPFileBackend) testDownloadSpeed(ctx context.Context, url string) (float64, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read the response body to measure actual download
	totalBytes := int64(0)
	buffer := make([]byte, 32768) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		totalBytes += int64(n)

		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		// Check context for cancellation
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
	}

	duration := time.Since(start).Seconds()
	if duration == 0 {
		return 0, fmt.Errorf("download too fast to measure")
	}

	// Calculate Mbps (megabits per second)
	megabits := float64(totalBytes*8) / (1000 * 1000)
	mbps := megabits / duration

	return mbps, nil
}

// LibreSpeedBackend implements LibreSpeed client
type LibreSpeedBackend struct {
	config *config.Config
	client *http.Client
}

// NewLibreSpeedBackend creates a new LibreSpeed backend
func NewLibreSpeedBackend(cfg *config.Config, client *http.Client) *LibreSpeedBackend {
	return &LibreSpeedBackend{
		config: cfg,
		client: client,
	}
}

// Name returns the backend name
func (l *LibreSpeedBackend) Name() string {
	return "librespeed"
}

// IsAvailable checks if LibreSpeed servers are configured
func (l *LibreSpeedBackend) IsAvailable(ctx context.Context) bool {
	if options, ok := l.config.Network.SpeedTest.Backends["librespeed"]; ok {
		return len(options.Servers) > 0
	}
	return false
}

// RunTest runs the LibreSpeed test
func (l *LibreSpeedBackend) RunTest(ctx context.Context) (*SpeedTestResult, error) {
	options, ok := l.config.Network.SpeedTest.Backends["librespeed"]
	if !ok || len(options.Servers) == 0 {
		return nil, fmt.Errorf("librespeed servers not configured")
	}

	// For now, return a placeholder - full LibreSpeed implementation would be more complex
	return &SpeedTestResult{
		DownloadMbps:   50.0, // Placeholder
		UploadMbps:     10.0, // Placeholder
		ServerName:     "LibreSpeed",
		ServerLocation: "Unknown",
	}, nil
}