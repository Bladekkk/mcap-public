import os
import subprocess
import threading
import time
from pathlib import Path
import psutil
from datetime import datetime

class ServerManager:
    def __init__(self):
        self.servers = {}
        self.lock = threading.Lock()
    
    def start_server(self, server_id, jar_path, java_path, xms_mb, xmx_mb):
        server_dir = Path("/home/bladekkk/MCAP-project/data/servers") / server_id
        
        java_args = [
            java_path,
            f"-Xms{xms_mb}M",
            f"-Xmx{xmx_mb}M",
            "-XX:+UseG1GC",
            "-jar", jar_path,
            "nogui"
        ]
        
        process = subprocess.Popen(
            java_args,
            cwd=str(server_dir),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1
        )
        
        with self.lock:
            self.servers[server_id] = {
                "process": process,
                "status": "running",
                "start_time": datetime.now().isoformat(),
                "jar_path": jar_path,
                "java_path": java_path,
                "xms_mb": xms_mb,
                "xmx_mb": xmx_mb
            }
        
        return process
    
    def stop_server(self, server_id):
        with self.lock:
            if server_id not in self.servers:
                return {"error": "Server not found"}
            
            server = self.servers[server_id]
            process = server.get("process")
            
            if process and process.poll() is None:
                try:
                    process.terminate()
                    process.wait(timeout=30)
                except subprocess.TimeoutExpired:
                    process.kill()
            
            server["status"] = "stopped"
            return {"status": "stopped"}
    
    def restart_server(self, server_id):
        self.stop_server(server_id)
        time.sleep(2)
        
        server = self.servers.get(server_id)
        if server:
            return self.start_server(
                server_id,
                server["jar_path"],
                server["java_path"],
                server["xms_mb"],
                server["xmx_mb"]
            )
        return None
    
    def get_metrics(self, server_id):
        with self.lock:
            if server_id not in self.servers:
                return {"error": "Server not found"}
            
            server = self.servers[server_id]
            process = server.get("process")
            
            if not process or process.poll() is not None:
                return {
                    "status": "stopped",
                    "tps": 0,
                    "ram_mb": 0,
                    "cpu_percent": 0,
                    "online_players": 0
                }
            
            try:
                proc = psutil.Process(process.pid)
                mem_info = proc.memory_info()
                cpu_percent = proc.cpu_percent(interval=1)
                
                ram_mb = int(mem_info.rss / (1024 * 1024))
                
                return {
                    "status": "running",
                    "tps": 20.0,
                    "ram_mb": ram_mb,
                    "cpu_percent": round(cpu_percent, 2),
                    "online_players": 0
                }
            except psutil.NoSuchProcess:
                return {
                    "status": "crashed",
                    "tps": 0,
                    "ram_mb": 0,
                    "cpu_percent": 0,
                    "online_players": 0
                }
