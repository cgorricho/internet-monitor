"""
Internet Connection Monitor - Configuration Management
Centralized configuration loading and validation using environment variables
"""

import os
from pathlib import Path
from datetime import time
from typing import Optional
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

class Config:
    """Application configuration loaded from environment variables."""
    
    def __init__(self):
        # Get project root directory
        self.PROJECT_ROOT = Path(__file__).parent.parent.absolute()
        
        # Database Configuration
        self.DATABASE_PATH = self._get_path_env('DATABASE_PATH', 'data/internet_monitor.db')
        self.BACKUP_DATABASE_PATH = self._get_path_env('BACKUP_DATABASE_PATH', 'data/backups/')
        
        # Logging Configuration
        self.LOG_PATH = self._get_path_env('LOG_PATH', 'logs/')
        self.LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO').upper()
        self.LOG_ROTATION = os.getenv('LOG_ROTATION', '10MB')
        
        # Peak Hours Configuration
        self.PEAK_HOURS_START = self._parse_time(os.getenv('PEAK_HOURS_START', '09:00'))
        self.PEAK_HOURS_END = self._parse_time(os.getenv('PEAK_HOURS_END', '18:00'))
        
        # Monitoring Intervals (in minutes)
        # Peak hours
        self.PEAK_QUICK_INTERVAL = int(os.getenv('PEAK_QUICK_INTERVAL', '2'))
        self.PEAK_SPEEDTEST_INTERVAL = int(os.getenv('PEAK_SPEEDTEST_INTERVAL', '10'))
        self.PEAK_ANALYSIS_INTERVAL = int(os.getenv('PEAK_ANALYSIS_INTERVAL', '30'))
        
        # Off-peak hours
        self.OFFPEAK_QUICK_INTERVAL = int(os.getenv('OFFPEAK_QUICK_INTERVAL', '5'))
        self.OFFPEAK_SPEEDTEST_INTERVAL = int(os.getenv('OFFPEAK_SPEEDTEST_INTERVAL', '20'))
        self.OFFPEAK_ANALYSIS_INTERVAL = int(os.getenv('OFFPEAK_ANALYSIS_INTERVAL', '60'))
        
        # Alert Thresholds
        self.ALERT_DOWNLOAD_WARNING = float(os.getenv('ALERT_DOWNLOAD_WARNING', '25'))
        self.ALERT_DOWNLOAD_CRITICAL = float(os.getenv('ALERT_DOWNLOAD_CRITICAL', '10'))
        
        self.ALERT_UPLOAD_WARNING = float(os.getenv('ALERT_UPLOAD_WARNING', '5'))
        self.ALERT_UPLOAD_CRITICAL = float(os.getenv('ALERT_UPLOAD_CRITICAL', '1'))
        
        self.ALERT_PING_WARNING = float(os.getenv('ALERT_PING_WARNING', '50'))
        self.ALERT_PING_CRITICAL = float(os.getenv('ALERT_PING_CRITICAL', '100'))
        
        self.ALERT_PACKET_LOSS_WARNING = float(os.getenv('ALERT_PACKET_LOSS_WARNING', '1'))
        self.ALERT_PACKET_LOSS_CRITICAL = float(os.getenv('ALERT_PACKET_LOSS_CRITICAL', '3'))
        
        self.ALERT_WIFI_WARNING = float(os.getenv('ALERT_WIFI_WARNING', '-70'))
        self.ALERT_WIFI_CRITICAL = float(os.getenv('ALERT_WIFI_CRITICAL', '-80'))
        
        self.ALERT_DNS_WARNING = float(os.getenv('ALERT_DNS_WARNING', '100'))
        self.ALERT_DNS_CRITICAL = float(os.getenv('ALERT_DNS_CRITICAL', '500'))
        
        # Data Retention
        self.DATA_RETENTION_DAYS = int(os.getenv('DATA_RETENTION_DAYS', '90'))
        
        # Dashboard Configuration
        self.DASHBOARD_HOST = os.getenv('DASHBOARD_HOST', '127.0.0.1')
        self.DASHBOARD_PORT = int(os.getenv('DASHBOARD_PORT', '8050'))
        self.DASHBOARD_DEBUG = os.getenv('DASHBOARD_DEBUG', 'False').lower() == 'true'
        self.DASHBOARD_AUTO_RELOAD = os.getenv('DASHBOARD_AUTO_RELOAD', 'True').lower() == 'true'
        self.DASHBOARD_UPDATE_INTERVAL = int(os.getenv('DASHBOARD_UPDATE_INTERVAL', '60'))
        
        # Network Test Configuration
        self.SPEEDTEST_TIMEOUT = int(os.getenv('SPEEDTEST_TIMEOUT', '60'))
        self.PING_COUNT = int(os.getenv('PING_COUNT', '4'))
        self.DNS_TEST_DOMAIN = os.getenv('DNS_TEST_DOMAIN', 'google.com')
        self.PING_HOST = os.getenv('PING_HOST', '8.8.8.8')
        
        # Application Settings
        self.APP_TIMEZONE = os.getenv('APP_TIMEZONE', 'America/New_York')
        self.MAX_RETRIES = int(os.getenv('MAX_RETRIES', '3'))
        self.RETRY_DELAY = int(os.getenv('RETRY_DELAY', '5'))
        
        # Ensure required directories exist
        self._create_directories()
        
    def _get_path_env(self, env_name: str, default: str) -> Path:
        """Get path from environment variable, resolve relative to project root."""
        path_str = os.getenv(env_name, default)
        path = Path(path_str)
        
        # If relative path, make it relative to project root
        if not path.is_absolute():
            path = self.PROJECT_ROOT / path
            
        return path
    
    def _parse_time(self, time_str: str) -> time:
        """Parse time string in HH:MM format."""
        try:
            hour, minute = map(int, time_str.split(':'))
            return time(hour, minute)
        except (ValueError, AttributeError):
            raise ValueError(f"Invalid time format: {time_str}. Expected HH:MM")
    
    def _create_directories(self):
        """Create required directories if they don't exist."""
        directories = [
            self.DATABASE_PATH.parent,
            self.BACKUP_DATABASE_PATH,
            self.LOG_PATH
        ]
        
        for directory in directories:
            directory.mkdir(parents=True, exist_ok=True)
    
    def is_peak_hours(self, current_time: Optional[time] = None) -> bool:
        """Check if current time is within peak hours."""
        if current_time is None:
            from datetime import datetime
            current_time = datetime.now().time()
            
        return self.PEAK_HOURS_START <= current_time <= self.PEAK_HOURS_END
    
    def get_alert_level(self, measurement_type: str, value: float) -> str:
        """
        Determine alert level based on measurement value and thresholds.
        
        Args:
            measurement_type: Type of measurement (download, upload, ping, etc.)
            value: Measured value
            
        Returns:
            'normal', 'warning', or 'critical'
        """
        thresholds = self._get_thresholds(measurement_type)
        if not thresholds:
            return 'normal'
            
        warning, critical = thresholds
        
        # For metrics where higher is worse (ping, packet_loss, dns)
        if measurement_type in ['ping', 'packet_loss', 'dns']:
            if value >= critical:
                return 'critical'
            elif value >= warning:
                return 'warning'
        # For metrics where lower is worse (download, upload, wifi)
        elif measurement_type in ['download', 'upload', 'wifi']:
            if value <= critical:
                return 'critical'
            elif value <= warning:
                return 'warning'
                
        return 'normal'
    
    def _get_thresholds(self, measurement_type: str) -> Optional[tuple]:
        """Get warning and critical thresholds for measurement type."""
        threshold_map = {
            'download': (self.ALERT_DOWNLOAD_WARNING, self.ALERT_DOWNLOAD_CRITICAL),
            'upload': (self.ALERT_UPLOAD_WARNING, self.ALERT_UPLOAD_CRITICAL),
            'ping': (self.ALERT_PING_WARNING, self.ALERT_PING_CRITICAL),
            'packet_loss': (self.ALERT_PACKET_LOSS_WARNING, self.ALERT_PACKET_LOSS_CRITICAL),
            'wifi': (self.ALERT_WIFI_WARNING, self.ALERT_WIFI_CRITICAL),
            'dns': (self.ALERT_DNS_WARNING, self.ALERT_DNS_CRITICAL)
        }
        
        return threshold_map.get(measurement_type)
    
    def get_monitoring_intervals(self, is_peak: Optional[bool] = None) -> dict:
        """Get monitoring intervals based on peak/off-peak status."""
        if is_peak is None:
            is_peak = self.is_peak_hours()
            
        if is_peak:
            return {
                'quick': self.PEAK_QUICK_INTERVAL,
                'speedtest': self.PEAK_SPEEDTEST_INTERVAL,
                'analysis': self.PEAK_ANALYSIS_INTERVAL
            }
        else:
            return {
                'quick': self.OFFPEAK_QUICK_INTERVAL,
                'speedtest': self.OFFPEAK_SPEEDTEST_INTERVAL,
                'analysis': self.OFFPEAK_ANALYSIS_INTERVAL
            }
    
    def __repr__(self) -> str:
        """String representation of configuration (excluding sensitive data)."""
        return f"Config(database={self.DATABASE_PATH}, peak_hours={self.PEAK_HOURS_START}-{self.PEAK_HOURS_END})"


# Global configuration instance
config = Config()

# Convenience constants for common paths
DB_PATH = config.DATABASE_PATH
LOG_PATH = config.LOG_PATH
PROJECT_ROOT = config.PROJECT_ROOT