import urllib.request
import urllib.parse
import json
from datetime import datetime, timedelta

class PluginAggregator:
    def __init__(self):
        self.cache = {}
        self.cache_ttl = timedelta(minutes=10)
    
    def search(self, query, loaders=None, versions=None):
        results = []
        
        try:
            modrinth_results = self._search_modrinth(query, versions)
            results.extend(modrinth_results)
        except Exception as e:
            print(f"Modrinth search error: {e}")
        
        try:
            hangar_results = self._search_hangar(query)
            results.extend(hangar_results)
        except Exception as e:
            print(f"Hangar search error: {e}")
        
        return results
    
    def _search_modrinth(self, query, versions):
        base_url = "https://api.modrinth.com/v2/search"
        
        params = {"query": query}
        if versions:
            facets = ["&".join([f'versions:{v}' for v in versions])]
            params["facets"] = facets
        
        url = f"{base_url}?{urllib.parse.urlencode(params)}"
        
        with urllib.request.urlopen(url, timeout=10) as response:
            data = json.loads(response.read().decode())
        
        plugins = []
        for hit in data.get("hits", []):
            files = hit.get("files", [])
            plugins.append({
                "id": hit.get("project_id"),
                "name": hit.get("title"),
                "description": hit.get("description"),
                "author": hit.get("author"),
                "downloads": hit.get("downloads", 0),
                "source": "modrinth",
                "download_url": files[0].get("url") if files else None
            })
        
        return plugins
    
    def _search_hangar(self, query):
        base_url = "https://hangar.papermc.io/api/v1/projects"
        
        if query:
            url = f"{base_url}?q={urllib.parse.quote(query)}"
        else:
            url = base_url
        
        with urllib.request.urlopen(url, timeout=10) as response:
            data = json.loads(response.read().decode())
        
        plugins = []
        for project in data:
            plugins.append({
                "id": project.get("slug"),
                "name": project.get("name"),
                "description": project.get("description"),
                "author": project.get("author"),
                "downloads": project.get("downloads", 0),
                "source": "hangar",
                "website": project.get("links", {}).get("website")
            })
        
        return plugins
