"""
Internet Connection Monitor - Network Testing Module
Comprehensive network performance measurement functions
"""

import json
import subprocess
import time
import re
import socket
from datetime import datetime
from typing import Dict, Optional, Any, List
from pathlib import Path
import requests
from ping3 import ping

from .config import config


class NetworkTester:
    """Handles all network performance testing operations."""
    
    def __init__(self):
        self.last_public_ip = None
        self.consecutive_failures = 0
        self.max_retries = config.MAX_RETRIES
        self.retry_delay = config.RETRY_DELAY
        
    def run_comprehensive_test(self) -> Dict[str, Any]:
        """
        Run a complete set of network tests.
        
        Returns:
            Dictionary with all network metrics
        """
        results = {
            'timestamp': datetime.now().isoformat(),
            'test_type': 'full'
        }
        
        # Run speed test (most comprehensive)
        speedtest_results = self.run_speedtest()
        results.update(speedtest_results)
        
        # Add additional metrics not covered by speedtest
        results['dns_resolution_ms'] = self.get_dns_resolution_time()
        results['wifi_signal_dbm'] = self.get_wifi_signal_strength()
        
        # Determine overall alert level
        results['alert_level'] = self._evaluate_alert_level(results)
        
        return results
    
    def run_quick_test(self) -> Dict[str, Any]:
        """
        Run lightweight network tests for frequent monitoring.
        
        Returns:
            Dictionary with basic network metrics
        """
        results = {
            'timestamp': datetime.now().isoformat(),
            'test_type': 'quick'
        }
        
        # Quick ping test
        results['ping_ms'] = self.quick_ping()
        
        # DNS resolution time
        results['dns_resolution_ms'] = self.get_dns_resolution_time()
        
        # WiFi signal strength (if available)
        results['wifi_signal_dbm'] = self.get_wifi_signal_strength()
        
        # Public IP (cached from last full test)
        if self.last_public_ip:
            results['public_ip'] = self.last_public_ip
        
        # Determine alert level based on available metrics
        results['alert_level'] = self._evaluate_alert_level(results)
        
        return results
    
    def run_speedtest(self) -> Dict[str, Any]:
        """
        Execute Ookla Speedtest CLI and parse results.
        
        Returns:
            Dictionary with speed test results
        """
        for attempt in range(self.max_retries):
            try:
                # Run speedtest with JSON output
                cmd = [
                    'speedtest', 
                    '--accept-license', 
                    '--accept-gdpr',
                    '--format=json',
                    f'--timeout={config.SPEEDTEST_TIMEOUT}'
                ]
                
                result = subprocess.run(
                    cmd,
                    capture_output=True,
                    text=True,
                    timeout=config.SPEEDTEST_TIMEOUT + 10
                )
                
                if result.returncode != 0:
                    raise subprocess.CalledProcessError(result.returncode, cmd, result.stderr)
                
                # Parse JSON output
                data = json.loads(result.stdout)
                
                # Extract key metrics
                speedtest_data = {
                    'download_mbps': round(data['download']['bandwidth'] * 8 / 1_000_000, 2),  # Convert bytes to Mbps
                    'upload_mbps': round(data['upload']['bandwidth'] * 8 / 1_000_000, 2),
                    'ping_ms': round(data['ping']['latency'], 2),
                    'jitter_ms': round(data['ping']['jitter'], 2),
                    'packet_loss_percent': data.get('packetLoss', 0),
                    'server_name': data['server']['name'],
                    'server_location': f"{data['server']['location']}, {data['server']['country']}",
                    'public_ip': data['interface']['externalIp']
                }
                
                # Cache public IP
                self.last_public_ip = speedtest_data['public_ip']
                self.consecutive_failures = 0  # Reset failure counter
                
                return speedtest_data
                
            except (subprocess.CalledProcessError, json.JSONDecodeError, KeyError, subprocess.TimeoutExpired) as e:
                self.consecutive_failures += 1
                
                if attempt < self.max_retries - 1:
                    print(f"⚠️  Speedtest attempt {attempt + 1} failed: {e}. Retrying in {self.retry_delay}s...")
                    time.sleep(self.retry_delay)
                    continue
                else:
                    print(f"❌ Speedtest failed after {self.max_retries} attempts: {e}")
                    return self._get_fallback_speedtest_data()
    
    def quick_ping(self, host: str = None) -> Optional[float]:
        """
        Perform a quick ping test to measure latency.
        
        Args:
            host: Target host (defaults to config.PING_HOST)
            
        Returns:
            Ping latency in milliseconds, None if failed
        """
        target_host = host or config.PING_HOST
        
        try:
            # Use ping3 library for consistent cross-platform behavior
            response_time = ping(target_host, timeout=5, unit='ms')
            
            if response_time is not None:
                return round(response_time, 2)
            else:
                # Fallback to system ping if ping3 fails
                return self._system_ping(target_host)
                
        except Exception as e:
            print(f"⚠️  Ping failed: {e}")
            return self._system_ping(target_host)
    
    def _system_ping(self, host: str) -> Optional[float]:
        """Fallback ping using system ping command."""
        try:
            result = subprocess.run(
                ['ping', '-c', str(config.PING_COUNT), '-W', '5', host],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            if result.returncode == 0:
                # Parse avg time from ping output
                # Example: "rtt min/avg/max/mdev = 12.345/23.456/34.567/5.678 ms"
                match = re.search(r'rtt min/avg/max/mdev = [^/]+/([^/]+)/', result.stdout)
                if match:
                    return round(float(match.group(1)), 2)
                    
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, ValueError) as e:
            print(f"⚠️  System ping failed: {e}")
            
        return None
    
    def get_dns_resolution_time(self, domain: str = None) -> Optional[float]:
        """
        Measure DNS resolution time.
        
        Args:
            domain: Domain to resolve (defaults to config.DNS_TEST_DOMAIN)
            
        Returns:
            DNS resolution time in milliseconds, None if failed
        """
        target_domain = domain or config.DNS_TEST_DOMAIN
        
        try:
            start_time = time.time()
            socket.gethostbyname(target_domain)
            end_time = time.time()
            
            resolution_time = (end_time - start_time) * 1000  # Convert to milliseconds
            return round(resolution_time, 2)
            
        except socket.gaierror as e:
            print(f"⚠️  DNS resolution failed for {target_domain}: {e}")
            return None
    
    def get_wifi_signal_strength(self) -> Optional[float]:
        """
        Get WiFi signal strength using iwconfig.
        
        Returns:
            Signal strength in dBm, None if not available or wired connection
        """
        try:
            # Run iwconfig to get wireless interface info
            result = subprocess.run(
                ['iwconfig'],
                capture_output=True,
                text=True,
                timeout=5
            )
            
            if result.returncode == 0:
                # Parse signal strength from iwconfig output
                # Example: "Link Quality=70/70  Signal level=-40 dBm"
                match = re.search(r'Signal level=(-?\d+(?:\.\d+)?)\s*dBm', result.stdout)
                if match:
                    return float(match.group(1))
                    
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError) as e:
            # iwconfig not available or no wireless interface
            pass
            
        return None
    
    def get_public_ip(self) -> Optional[str]:
        """
        Get current public IP address.
        
        Returns:
            Public IP address string, None if failed
        """
        # Try multiple IP detection services for reliability
        services = [
            'https://api.ipify.org',
            'https://icanhazip.com',
            'https://ifconfig.me/ip'
        ]
        
        for service in services:
            try:
                response = requests.get(service, timeout=10)
                if response.status_code == 200:
                    ip = response.text.strip()
                    # Basic IP validation
                    if re.match(r'^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$', ip):
                        self.last_public_ip = ip
                        return ip
                        
            except requests.RequestException:
                continue
                
        return self.last_public_ip  # Return cached IP if all services fail
    
    def _get_fallback_speedtest_data(self) -> Dict[str, Any]:
        """Return fallback data when speedtest fails completely."""
        return {
            'download_mbps': None,
            'upload_mbps': None,
            'ping_ms': self.quick_ping(),  # Try to get at least ping
            'jitter_ms': None,
            'packet_loss_percent': None,
            'server_name': 'Unknown',
            'server_location': 'Unknown',
            'public_ip': self.get_public_ip()
        }
    
    def _evaluate_alert_level(self, results: Dict[str, Any]) -> str:
        """
        Evaluate overall alert level based on all available metrics.
        
        Args:
            results: Dictionary with measurement results
            
        Returns:
            Overall alert level ('normal', 'warning', 'critical')
        """
        alert_levels = []
        
        # Check each metric against thresholds
        metrics_to_check = {
            'download_mbps': 'download',
            'upload_mbps': 'upload', 
            'ping_ms': 'ping',
            'packet_loss_percent': 'packet_loss',
            'wifi_signal_dbm': 'wifi',
            'dns_resolution_ms': 'dns'
        }
        
        for result_key, config_key in metrics_to_check.items():
            value = results.get(result_key)
            if value is not None:
                alert_level = config.get_alert_level(config_key, value)
                alert_levels.append(alert_level)
        
        # Return highest priority alert level
        if 'critical' in alert_levels:
            return 'critical'
        elif 'warning' in alert_levels:
            return 'warning'
        else:
            return 'normal'
    
    def diagnose_connection(self) -> Dict[str, Any]:
        """
        Run comprehensive connection diagnostics.
        
        Returns:
            Dictionary with diagnostic information
        """
        diagnostics = {
            'timestamp': datetime.now().isoformat(),
            'tests': {}
        }
        
        # Basic connectivity test
        diagnostics['tests']['ping_google'] = self.quick_ping('8.8.8.8')
        diagnostics['tests']['ping_cloudflare'] = self.quick_ping('1.1.1.1')
        
        # DNS resolution tests
        diagnostics['tests']['dns_google'] = self.get_dns_resolution_time('google.com')
        diagnostics['tests']['dns_cloudflare'] = self.get_dns_resolution_time('cloudflare.com')
        
        # Public IP detection
        diagnostics['tests']['public_ip'] = self.get_public_ip()
        
        # WiFi signal strength
        diagnostics['tests']['wifi_signal'] = self.get_wifi_signal_strength()
        
        # Connection stability (multiple pings)
        ping_results = []
        for _ in range(5):
            ping_result = self.quick_ping()
            if ping_result:
                ping_results.append(ping_result)
            time.sleep(1)
        
        if ping_results:
            diagnostics['tests']['ping_stability'] = {
                'min': min(ping_results),
                'max': max(ping_results),
                'avg': sum(ping_results) / len(ping_results),
                'jitter': max(ping_results) - min(ping_results)
            }
        
        return diagnostics
    
    def get_connection_status(self) -> Dict[str, Any]:
        """
        Get current connection status summary.
        
        Returns:
            Dictionary with connection status
        """
        status = {
            'timestamp': datetime.now().isoformat(),
            'connected': False,
            'internet_access': False,
            'dns_working': False,
            'wifi_connected': False
        }
        
        # Test basic connectivity
        ping_result = self.quick_ping()
        if ping_result:
            status['connected'] = True
            status['latency_ms'] = ping_result
        
        # Test internet access
        public_ip = self.get_public_ip()
        if public_ip:
            status['internet_access'] = True
            status['public_ip'] = public_ip
        
        # Test DNS
        dns_time = self.get_dns_resolution_time()
        if dns_time:
            status['dns_working'] = True
            status['dns_latency_ms'] = dns_time
        
        # Check WiFi
        wifi_signal = self.get_wifi_signal_strength()
        if wifi_signal:
            status['wifi_connected'] = True
            status['wifi_signal_dbm'] = wifi_signal
        
        return status


# Global network tester instance
network_tester = NetworkTester()

# Convenience functions for common operations
def run_comprehensive_test() -> Dict[str, Any]:
    """Run comprehensive network test."""
    return network_tester.run_comprehensive_test()

def run_quick_test() -> Dict[str, Any]:
    """Run quick network test."""
    return network_tester.run_quick_test()

def quick_ping(host: str = None) -> Optional[float]:
    """Quick ping test."""
    return network_tester.quick_ping(host)

def get_connection_status() -> Dict[str, Any]:
    """Get current connection status."""
    return network_tester.get_connection_status()