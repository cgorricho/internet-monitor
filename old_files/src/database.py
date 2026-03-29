"""
Internet Connection Monitor - Database Layer
SQLite database operations for network performance measurements
"""

import sqlite3
import sqlite_utils
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from pathlib import Path
import json

from .config import config


class DatabaseManager:
    """Manages SQLite database operations for network monitoring data."""
    
    def __init__(self, db_path: Optional[Path] = None):
        """Initialize database manager with configurable path."""
        self.db_path = db_path or config.DATABASE_PATH
        self.db = sqlite_utils.Database(self.db_path)
        
    def init_db(self) -> None:
        """Initialize database schema and indexes."""
        # Create measurements table with comprehensive schema
        self.db.executescript("""
            CREATE TABLE IF NOT EXISTS measurements (
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
                alert_level TEXT DEFAULT 'normal',
                test_type TEXT DEFAULT 'full',
                server_name TEXT,
                server_location TEXT,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            );
            
            -- Create indexes for performance
            CREATE INDEX IF NOT EXISTS idx_measurements_timestamp ON measurements(timestamp);
            CREATE INDEX IF NOT EXISTS idx_measurements_alert_level ON measurements(alert_level);
            CREATE INDEX IF NOT EXISTS idx_measurements_test_type ON measurements(test_type);
            CREATE INDEX IF NOT EXISTS idx_measurements_created_at ON measurements(created_at);
            
            -- Create alerts table for detailed alert tracking
            CREATE TABLE IF NOT EXISTS alerts (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                measurement_id INTEGER,
                alert_type TEXT NOT NULL,
                alert_level TEXT NOT NULL,
                metric_name TEXT NOT NULL,
                metric_value REAL NOT NULL,
                threshold_value REAL NOT NULL,
                message TEXT,
                resolved BOOLEAN DEFAULT FALSE,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                resolved_at DATETIME,
                FOREIGN KEY (measurement_id) REFERENCES measurements (id)
            );
            
            CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
            CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved);
            CREATE INDEX IF NOT EXISTS idx_alerts_level ON alerts(alert_level);
            
            -- Create performance summary table for quick dashboard queries
            CREATE TABLE IF NOT EXISTS performance_summary (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                date DATE NOT NULL,
                hour INTEGER NOT NULL,
                avg_download_mbps REAL,
                avg_upload_mbps REAL,
                avg_ping_ms REAL,
                avg_jitter_ms REAL,
                avg_packet_loss_percent REAL,
                avg_dns_resolution_ms REAL,
                avg_wifi_signal_dbm REAL,
                measurement_count INTEGER DEFAULT 0,
                alert_count INTEGER DEFAULT 0,
                uptime_percent REAL DEFAULT 100.0,
                is_peak_hour BOOLEAN DEFAULT FALSE,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(date, hour)
            );
            
            CREATE INDEX IF NOT EXISTS idx_summary_date_hour ON performance_summary(date, hour);
            CREATE INDEX IF NOT EXISTS idx_summary_is_peak ON performance_summary(is_peak_hour);
        """)
        
        print(f"✅ Database initialized at {self.db_path}")
        
    def insert_measurement(self, data: Dict[str, Any]) -> int:
        """
        Insert a new measurement record.
        
        Args:
            data: Dictionary containing measurement data
            
        Returns:
            ID of inserted record
        """
        # Ensure timestamp is set
        if 'timestamp' not in data:
            data['timestamp'] = datetime.now().isoformat()
            
        # Insert measurement and return ID
        record_id = self.db['measurements'].insert(data).last_pk
        
        # Update hourly summary
        self._update_performance_summary(data)
        
        return record_id
    
    def fetch_measurements(
        self,
        hours: Optional[int] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        alert_level: Optional[str] = None,
        test_type: Optional[str] = None,
        limit: Optional[int] = None
    ) -> List[Dict]:
        """
        Fetch measurements with flexible filtering options.
        
        Args:
            hours: Number of hours back from now
            start_time: Start datetime for range query
            end_time: End datetime for range query  
            alert_level: Filter by alert level ('normal', 'warning', 'critical')
            test_type: Filter by test type ('quick', 'full')
            limit: Maximum number of records to return
            
        Returns:
            List of measurement dictionaries
        """
        query = "SELECT * FROM measurements WHERE 1=1"
        params = []
        
        # Time range filtering
        if hours:
            cutoff_time = datetime.now() - timedelta(hours=hours)
            query += " AND timestamp >= ?"
            params.append(cutoff_time.isoformat())
        
        if start_time:
            query += " AND timestamp >= ?"
            params.append(start_time.isoformat())
            
        if end_time:
            query += " AND timestamp <= ?"
            params.append(end_time.isoformat())
        
        # Filter by alert level
        if alert_level:
            query += " AND alert_level = ?"
            params.append(alert_level)
            
        # Filter by test type
        if test_type:
            query += " AND test_type = ?"
            params.append(test_type)
        
        # Order by timestamp (most recent first)
        query += " ORDER BY timestamp DESC"
        
        # Apply limit
        if limit:
            query += " LIMIT ?"
            params.append(limit)
            
        return list(self.db.execute(query, params))
    
    def get_latest_measurement(self) -> Optional[Dict]:
        """Get the most recent measurement."""
        measurements = self.fetch_measurements(limit=1)
        return measurements[0] if measurements else None
    
    def get_alerts(
        self,
        hours: Optional[int] = 24,
        resolved: Optional[bool] = None,
        alert_level: Optional[str] = None
    ) -> List[Dict]:
        """
        Fetch alerts with filtering options.
        
        Args:
            hours: Number of hours back from now (default 24)
            resolved: Filter by resolution status
            alert_level: Filter by alert level
            
        Returns:
            List of alert dictionaries
        """
        query = "SELECT * FROM alerts WHERE 1=1"
        params = []
        
        if hours:
            cutoff_time = datetime.now() - timedelta(hours=hours)
            query += " AND created_at >= ?"
            params.append(cutoff_time.isoformat())
            
        if resolved is not None:
            query += " AND resolved = ?"
            params.append(resolved)
            
        if alert_level:
            query += " AND alert_level = ?"
            params.append(alert_level)
            
        query += " ORDER BY created_at DESC"
        
        return list(self.db.execute(query, params))
    
    def insert_alert(
        self,
        measurement_id: int,
        alert_type: str,
        alert_level: str,
        metric_name: str,
        metric_value: float,
        threshold_value: float,
        message: Optional[str] = None
    ) -> int:
        """Insert a new alert record."""
        alert_data = {
            'measurement_id': measurement_id,
            'alert_type': alert_type,
            'alert_level': alert_level,
            'metric_name': metric_name,
            'metric_value': metric_value,
            'threshold_value': threshold_value,
            'message': message or f"{metric_name} {alert_level}: {metric_value} (threshold: {threshold_value})",
            'created_at': datetime.now().isoformat()
        }
        
        return self.db['alerts'].insert(alert_data).last_pk
    
    def get_performance_summary(
        self,
        days: int = 7,
        is_peak_hour: Optional[bool] = None
    ) -> List[Dict]:
        """
        Get hourly performance summary for dashboard analytics.
        
        Args:
            days: Number of days back from now
            is_peak_hour: Filter by peak/off-peak hours
            
        Returns:
            List of hourly summary dictionaries
        """
        query = """
            SELECT * FROM performance_summary 
            WHERE date >= date('now', '-{} days')
        """.format(days)
        
        params = []
        
        if is_peak_hour is not None:
            query += " AND is_peak_hour = ?"
            params.append(is_peak_hour)
            
        query += " ORDER BY date DESC, hour DESC"
        
        return list(self.db.execute(query, params))
    
    def _update_performance_summary(self, measurement: Dict[str, Any]) -> None:
        """Update hourly performance summary with new measurement."""
        timestamp = datetime.fromisoformat(measurement['timestamp'].replace('Z', '+00:00'))
        date_str = timestamp.date().isoformat()
        hour = timestamp.hour
        
        # Determine if this is peak hour
        is_peak = config.is_peak_hours(timestamp.time())
        
        # Get existing summary or create new one
        existing = list(self.db.execute(
            "SELECT * FROM performance_summary WHERE date = ? AND hour = ?",
            [date_str, hour]
        ))
        
        if existing:
            # Update existing summary with running average
            summary = existing[0]
            count = summary['measurement_count'] + 1
            
            # Calculate new averages
            def update_avg(current_avg, new_value):
                if current_avg is None or new_value is None:
                    return new_value or current_avg
                return ((current_avg * (count - 1)) + new_value) / count
            
            updated_data = {
                'avg_download_mbps': update_avg(summary['avg_download_mbps'], measurement.get('download_mbps')),
                'avg_upload_mbps': update_avg(summary['avg_upload_mbps'], measurement.get('upload_mbps')),
                'avg_ping_ms': update_avg(summary['avg_ping_ms'], measurement.get('ping_ms')),
                'avg_jitter_ms': update_avg(summary['avg_jitter_ms'], measurement.get('jitter_ms')),
                'avg_packet_loss_percent': update_avg(summary['avg_packet_loss_percent'], measurement.get('packet_loss_percent')),
                'avg_dns_resolution_ms': update_avg(summary['avg_dns_resolution_ms'], measurement.get('dns_resolution_ms')),
                'avg_wifi_signal_dbm': update_avg(summary['avg_wifi_signal_dbm'], measurement.get('wifi_signal_dbm')),
                'measurement_count': count
            }
            
            self.db.execute(
                """UPDATE performance_summary 
                   SET avg_download_mbps = ?, avg_upload_mbps = ?, avg_ping_ms = ?,
                       avg_jitter_ms = ?, avg_packet_loss_percent = ?, avg_dns_resolution_ms = ?,
                       avg_wifi_signal_dbm = ?, measurement_count = ?
                   WHERE date = ? AND hour = ?""",
                [*updated_data.values(), date_str, hour]
            )
        else:
            # Create new summary record
            summary_data = {
                'date': date_str,
                'hour': hour,
                'avg_download_mbps': measurement.get('download_mbps'),
                'avg_upload_mbps': measurement.get('upload_mbps'),
                'avg_ping_ms': measurement.get('ping_ms'),
                'avg_jitter_ms': measurement.get('jitter_ms'),
                'avg_packet_loss_percent': measurement.get('packet_loss_percent'),
                'avg_dns_resolution_ms': measurement.get('dns_resolution_ms'),
                'avg_wifi_signal_dbm': measurement.get('wifi_signal_dbm'),
                'measurement_count': 1,
                'alert_count': 1 if measurement.get('alert_level', 'normal') != 'normal' else 0,
                'is_peak_hour': is_peak
            }
            
            self.db['performance_summary'].insert(summary_data)
    
    def purge_old_data(self, days: int = None) -> int:
        """
        Remove old measurement records to maintain database size.
        
        Args:
            days: Number of days to retain (defaults to config value)
            
        Returns:
            Number of records deleted
        """
        retention_days = days or config.DATA_RETENTION_DAYS
        cutoff_date = datetime.now() - timedelta(days=retention_days)
        
        # Delete old measurements
        result = self.db.execute(
            "DELETE FROM measurements WHERE timestamp < ?",
            [cutoff_date.isoformat()]
        )
        
        deleted_measurements = result.rowcount
        
        # Delete old alerts
        self.db.execute(
            "DELETE FROM alerts WHERE created_at < ?",
            [cutoff_date.isoformat()]
        )
        
        # Delete old performance summaries
        self.db.execute(
            "DELETE FROM performance_summary WHERE date < ?",
            [cutoff_date.date().isoformat()]
        )
        
        # Run VACUUM to reclaim space
        self.db.execute("VACUUM")
        
        return deleted_measurements
    
    def get_database_stats(self) -> Dict[str, Any]:
        """Get database statistics for monitoring."""
        stats = {}
        
        # Table record counts
        stats['measurements_count'] = self.db.execute("SELECT COUNT(*) as count FROM measurements").fetchone()['count']
        stats['alerts_count'] = self.db.execute("SELECT COUNT(*) as count FROM alerts").fetchone()['count']
        stats['summary_count'] = self.db.execute("SELECT COUNT(*) as count FROM performance_summary").fetchone()['count']
        
        # Database file size
        if self.db_path.exists():
            stats['database_size_mb'] = round(self.db_path.stat().st_size / (1024 * 1024), 2)
        else:
            stats['database_size_mb'] = 0
            
        # Date range of data
        date_range = self.db.execute(
            "SELECT MIN(timestamp) as min_date, MAX(timestamp) as max_date FROM measurements"
        ).fetchone()
        
        stats['data_range'] = {
            'earliest': date_range['min_date'],
            'latest': date_range['max_date']
        }
        
        return stats
    
    def backup_database(self, backup_path: Optional[Path] = None) -> Path:
        """Create a backup of the database."""
        if backup_path is None:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            backup_path = config.BACKUP_DATABASE_PATH / f"internet_monitor_backup_{timestamp}.db"
        
        # Ensure backup directory exists
        backup_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Create backup using sqlite3
        source = sqlite3.connect(self.db_path)
        backup = sqlite3.connect(backup_path)
        
        source.backup(backup)
        
        source.close()
        backup.close()
        
        return backup_path


# Global database instance
db_manager = DatabaseManager()

# Convenience functions for common operations
def init_db() -> None:
    """Initialize the database schema."""
    db_manager.init_db()

def insert_measurement(data: Dict[str, Any]) -> int:
    """Insert a measurement record."""
    return db_manager.insert_measurement(data)

def fetch_measurements(**kwargs) -> List[Dict]:
    """Fetch measurements with filtering."""
    return db_manager.fetch_measurements(**kwargs)

def get_latest_measurement() -> Optional[Dict]:
    """Get the most recent measurement."""
    return db_manager.get_latest_measurement()

def purge_old_data(days: int = None) -> int:
    """Purge old data from database."""
    return db_manager.purge_old_data(days)