from flask import Blueprint, request, jsonify
from flask_socketio import emit, join_room, leave_room
from sqlalchemy import text
from models import Server, JDK, DownloadedJAR, MetricsHistory
from config import Config
from services.java_manager import JavaManager
from services.downloader import Downloader
from services.server_manager import ServerManager
from services.plugin_aggregator import PluginAggregator
from database_init import socketio, db
from datetime import datetime
import os
import uuid

api_bp = Blueprint("api", __name__)

java_mgr = JavaManager()
downloader = Downloader(Config.DATA_DIR / "cache")
server_mgr = ServerManager()

def generate_id():
    return str(uuid.uuid4())

@api_bp.route("/servers", methods=["GET"])
def list_servers():
    servers = Server.query.all()
    return jsonify([s.to_dict() for s in servers])

@api_bp.route("/servers", methods=["POST"])
def create_server():
    data = request.json
    
    if not data.get("name") or not data.get("jar_path"):
        return jsonify({"error": "name and jar_path are required"}), 400
    
    if data.get("xms_mb", 0) > data.get("xmx_mb", 0):
        return jsonify({"error": "xms_mb must be <= xmx_mb"}), 400
    
    server = Server(
        id=generate_id(),
        name=data["name"],
        jar_path=data["jar_path"],
        port=data.get("port", 25565),
        java_version=data.get("java_version"),
        xms_mb=data.get("xms_mb", 1024),
        xmx_mb=data.get("xmx_mb", 2048),
        autostart=data.get("autostart", False),
        eula_accepted=data.get("eula_accepted", False),
        status="stopped"
    )
    
    db.session.add(server)
    db.session.commit()
    
    server_dir = Config.DATA_DIR / "servers" / server.id
    server_dir.mkdir(parents=True, exist_ok=True)
    
    (server_dir / "eula.txt").write_text(f"eula={server.eula_accepted}\n")
    
    return jsonify(server.to_dict()), 201

@api_bp.route("/servers/<server_id>", methods=["GET"])
def get_server(server_id):
    server = Server.query.get_or_404(server_id)
    return jsonify(server.to_dict())

@api_bp.route("/servers/<server_id>/start", methods=["POST"])
def start_server(server_id):
    server = Server.query.get_or_404(server_id)
    
    jar_path = Config.DATA_DIR / "servers" / server_id / os.path.basename(server.jar_path)
    if not jar_path.exists():
        return jsonify({"error": "jar file not found"}), 404
    
    java_path = server.java_version or "java"
    
    server_mgr.start_server(server_id, str(jar_path), java_path, server.xms_mb, server.xmx_mb)
    server.status = "running"
    db.session.commit()
    
    return jsonify({"status": "started", "id": server_id})

@api_bp.route("/servers/<server_id>/stop", methods=["POST"])
def stop_server(server_id):
    server = Server.query.get_or_404(server_id)
    
    server_mgr.stop_server(server_id)
    server.status = "stopped"
    db.session.commit()
    
    return jsonify({"status": "stopped", "id": server_id})

@api_bp.route("/servers/<server_id>/restart", methods=["POST"])
def restart_server(server_id):
    server = Server.query.get_or_404(server_id)
    
    server_mgr.restart_server(server_id)
    server.status = "running"
    db.session.commit()
    
    return jsonify({"status": "restarted", "id": server_id})

@api_bp.route("/servers/<server_id>", methods=["DELETE"])
def delete_server(server_id):
    server = Server.query.get_or_404(server_id)
    
    if server.status in ["running", "starting"]:
        server_mgr.stop_server(server_id)
    
    db.session.delete(server)
    db.session.commit()
    
    server_dir = Config.DATA_DIR / "servers" / server_id
    if server_dir.exists():
        import shutil
        shutil.rmtree(server_dir)
    
    return jsonify({"status": "deleted", "id": server_id})

@api_bp.route("/java", methods=["GET"])
def list_jdks():
    jdks = JDK.query.all()
    return jsonify([j.to_dict() for j in jdks])

@api_bp.route("/java", methods=["POST"])
def install_jdk():
    data = request.json
    version = data.get("version")
    
    if not version:
        return jsonify({"error": "version is required"}), 400
    
    os_name, arch = java_mgr.detect_system()
    java_path = java_mgr.download_jdk(version, os_name, arch)
    
    jdk = JDK(
        id=generate_id(),
        version=version,
        arch=arch,
        os=os_name,
        install_path=java_path
    )
    db.session.add(jdk)
    db.session.commit()
    
    return jsonify({"status": "installed", "version": version}), 201

@api_bp.route("/downloads", methods=["POST"])
def download_jar():
    data = request.json
    
    if not data.get("url") or not data.get("server_id"):
        return jsonify({"error": "url and server_id are required"}), 400
    
    filename = os.path.basename(data["url"])
    result = downloader.download(data["url"], filename)
    
    if result.get("error"):
        return jsonify(result), 400
    
    jar = DownloadedJAR(
        id=generate_id(),
        server_id=data["server_id"],
        filename=filename,
        url=data["url"],
        checksum=result.get("checksum"),
        size_bytes=result.get("size", 0)
    )
    db.session.add(jar)
    db.session.commit()
    
    return jsonify(result), 201

@api_bp.route("/metrics/<server_id>", methods=["GET"])
def get_metrics(server_id):
    server = Server.query.get_or_404(server_id)
    metrics = server_mgr.get_metrics(server_id)
    return jsonify(metrics)

@api_bp.route("/metrics/<server_id>/history", methods=["GET"])
def get_metrics_history(server_id):
    server = Server.query.get_or_404(server_id)
    
    history = MetricsHistory.query.filter_by(server_id=server_id)\
        .order_by(MetricsHistory.timestamp.desc())\
        .limit(100)\
        .all()
    
    return jsonify({
        "history": [h.to_dict() for h in history]
    })

@api_bp.route("/plugins", methods=["GET"])
def search_plugins():
    query = request.args.get("q", "")
    loaders = request.args.getlist("loaders")
    versions = request.args.getlist("versions")
    
    aggregator = PluginAggregator()
    plugins = aggregator.search(query, loaders, versions)
    
    return jsonify({"plugins": plugins})

@api_bp.route("/plugins/download", methods=["POST"])
def download_plugin():
    data = request.json
    
    if not data.get("plugin_id") or not data.get("server_id"):
        return jsonify({"error": "plugin_id and server_id are required"}), 400
    
    return jsonify({"error": "not implemented yet"}), 501

# WebSocket events
@socketio.on("connect")
def handle_connect():
    print("Client connected")

@socketio.on("disconnect")
def handle_disconnect():
    print("Client disconnected")

@socketio.on("join")
def handle_join(data):
    server_id = data.get("server_id")
    if server_id:
        join_room(server_id)
        emit("connected", {"server_id": server_id})

@socketio.on("leave")
def handle_leave(data):
    server_id = data.get("server_id")
    if server_id:
        leave_room(server_id)
