import hashlib
from pathlib import Path
import urllib.request

class Downloader:
    def __init__(self, cache_dir):
        self.cache_dir = Path(cache_dir)
        self.cache_dir.mkdir(parents=True, exist_ok=True)
    
    def download(self, url, filename):
        cache_path = self.cache_dir / filename
        
        if cache_path.exists():
            checksum = self._calculate_checksum(cache_path)
            return {"path": str(cache_path), "checksum": checksum}
        
        temp_path = self.cache_dir / f"{filename}.tmp"
        
        try:
            urllib.request.urlretrieve(url, temp_path)
            temp_path.rename(cache_path)
            
            checksum = self._calculate_checksum(cache_path)
            size = cache_path.stat().st_size
            
            return {"path": str(cache_path), "checksum": checksum, "size": size}
        except Exception as e:
            if temp_path.exists():
                temp_path.unlink()
            return {"error": str(e)}
    
    def _calculate_checksum(self, path):
        h = hashlib.sha1()
        with open(path, "rb") as f:
            for chunk in iter(lambda: f.read(8192), b""):
                h.update(chunk)
        return h.hexdigest()
