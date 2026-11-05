from datetime import datetime
from database_init import db

class Server(db.Model):
    __tablename__ = "servers"
    
    id = db.Column(db.String(36), primary_key=True)
    name = db.Column(db.String(50), nullable=False)
    jar_path = db.Column(db.String(255), nullable=False)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    status = db.Column(db.String(20), default="stopped")
    port = db.Column(db.Integer, default=25565)
    java_version = db.Column(db.String(20))
    xms_mb = db.Column(db.Integer, default=1024)
    xmx_mb = db.Column(db.Integer, default=2048)
    autostart = db.Column(db.Boolean, default=False)
    eula_accepted = db.Column(db.Boolean, default=False)
    
    metrics = db.relationship("MetricsHistory", backref="server", lazy=True, cascade="all, delete-orphan")
    
    def to_dict(self):
        return {
            "id": self.id,
            "name": self.name,
            "jar_path": self.jar_path,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "status": self.status,
            "port": self.port,
            "java_version": self.java_version,
            "xms_mb": self.xms_mb,
            "xmx_mb": self.xmx_mb,
            "autostart": self.autostart,
            "eula_accepted": self.eula_accepted
        }

class JDK(db.Model):
    __tablename__ = "jdk_versions"
    
    id = db.Column(db.String(36), primary_key=True)
    version = db.Column(db.String(20), nullable=False)
    arch = db.Column(db.String(20), nullable=False)
    os = db.Column(db.String(20), nullable=False)
    install_path = db.Column(db.String(255), nullable=False)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    def to_dict(self):
        return {
            "id": self.id,
            "version": self.version,
            "arch": self.arch,
            "os": self.os,
            "install_path": self.install_path,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class DownloadedJAR(db.Model):
    __tablename__ = "downloaded_jars"
    
    id = db.Column(db.String(36), primary_key=True)
    server_id = db.Column(db.String(36), db.ForeignKey("servers.id", ondelete="CASCADE"), nullable=False)
    filename = db.Column(db.String(255), nullable=False)
    url = db.Column(db.String(500), nullable=False)
    checksum = db.Column(db.String(128))
    size_bytes = db.Column(db.Integer)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    def to_dict(self):
        return {
            "id": self.id,
            "server_id": self.server_id,
            "filename": self.filename,
            "url": self.url,
            "checksum": self.checksum,
            "size_bytes": self.size_bytes,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class MetricsHistory(db.Model):
    __tablename__ = "metrics_history"
    
    id = db.Column(db.Integer, primary_key=True)
    server_id = db.Column(db.String(36), db.ForeignKey("servers.id", ondelete="CASCADE"), nullable=False)
    timestamp = db.Column(db.DateTime, default=datetime.utcnow)
    tps = db.Column(db.Float)
    mspt = db.Column(db.Float)
    online_players = db.Column(db.Integer)
    ram_mb = db.Column(db.Integer)
    cpu_percent = db.Column(db.Float)
    
    def to_dict(self):
        return {
            "id": self.id,
            "server_id": self.server_id,
            "timestamp": self.timestamp.isoformat() if self.timestamp else None,
            "tps": self.tps,
            "mspt": self.mspt,
            "online_players": self.online_players,
            "ram_mb": self.ram_mb,
            "cpu_percent": self.cpu_percent
        }
