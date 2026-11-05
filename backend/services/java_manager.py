import os
import platform
import urllib.request
import json
from pathlib import Path

class JavaManager:
    def __init__(self):
        self.base_dir = Path("/home/bladekkk/MCAP-project/data/jdk")
        self.base_dir.mkdir(parents=True, exist_ok=True)
    
    def detect_system(self):
        system = platform.system().lower()
        machine = platform.machine().lower()
        
        if system == "linux":
            os_name = "linux"
        elif system == "darwin":
            os_name = "macos"
        elif system == "windows":
            os_name = "windows"
        else:
            os_name = system
        
        if machine in ["x86_64", "amd64"]:
            arch = "x64"
        elif machine in ["aarch64", "arm64"]:
            arch = "aarch64"
        else:
            arch = machine
        
        return os_name, arch
    
    def get_java_version(self, version_str):
        try:
            parts = version_str.replace("forge-", "").replace("fabric-", "").split(".")
            major = int(parts[0])
            minor = int(parts[1]) if len(parts) > 1 else 0
            
            if major == 1:
                if minor <= 12:
                    return 8
                elif minor <= 16:
                    return 11
                elif minor <= 20:
                    return 17
                else:
                    return 21
            return 17
        except (ValueError, IndexError):
            return 17
    
    def download_jdk(self, version, os_name, arch):
        url = f"https://api.adoptium.net/v3/binary/latest/{version}/ga/{os_name}/{arch}/jdk/hotspot/normal/eclipse"
        
        try:
            with urllib.request.urlopen(url) as response:
                data = json.loads(response.read().decode())
            
            binary = data.get("binary", {})
            download_url = binary.get("link", "")
            filename = binary.get("name", f"jdk-{version}.tar.gz")
            
            if not download_url:
                raise Exception("No download URL found in Adoptium API response")
            
            temp_path = self.base_dir / f"jdk-{version}.tar.gz"
            urllib.request.urlretrieve(download_url, temp_path)
            
            return str(temp_path)
        except Exception as e:
            raise Exception(f"Failed to download JDK: {e}")
    
    def get_installed_jdks(self):
        jdks = []
        if self.base_dir.exists():
            for item in self.base_dir.iterdir():
                if item.is_file() and item.suffix == ".tar.gz":
                    jdks.append({
                        "path": str(item),
                        "size": item.stat().st_size
                    })
        return jdks
