"use client";

import { useState, useEffect } from "react";

interface Server {
  id: string;
  name: string;
  status: "running" | "stopped" | "starting";
  port: number;
  players?: number;
  ram_mb?: number;
}

export default function Home() {
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchServers();
    const interval = setInterval(fetchServers, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchServers = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/v1/servers");
      if (!response.ok) throw new Error("Failed to fetch servers");
      const data = await response.json();
      setServers(Array.isArray(data) ? data : []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  };

  const startServer = async (serverId: string) => {
    try {
      const response = await fetch(
        `http://localhost:8080/api/v1/servers/${serverId}/start`,
        { method: "POST" }
      );
      if (!response.ok) throw new Error("Failed to start server");
      await fetchServers();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  };

  const stopServer = async (serverId: string) => {
    try {
      const response = await fetch(
        `http://localhost:8080/api/v1/servers/${serverId}/stop`,
        { method: "POST" }
      );
      if (!response.ok) throw new Error("Failed to stop server");
      await fetchServers();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800">
      <header className="bg-slate-950 border-b border-slate-700 shadow-lg">
        <div className="max-w-7xl mx-auto px-6 py-6">
          <h1 className="text-3xl font-bold text-white">
            🎮 MCAP - Minecraft Admin Panel
          </h1>
          <p className="text-slate-400 mt-1">Server Management Dashboard</p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {error && (
          <div className="bg-red-900/30 border border-red-700 text-red-200 px-4 py-3 rounded-lg mb-6">
            ⚠️ {error}
          </div>
        )}

        {loading ? (
          <div className="text-center py-12">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
            <p className="text-slate-300 mt-4">Loading servers...</p>
          </div>
        ) : servers.length === 0 ? (
          <div className="bg-slate-700/50 border-2 border-dashed border-slate-600 rounded-lg p-12 text-center">
            <p className="text-slate-400 text-lg">No servers created yet.</p>
            <button className="mt-4 bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-lg transition">
              + Create New Server
            </button>
          </div>
        ) : (
          <div className="grid gap-6">
            {servers.map((server) => (
              <div
                key={server.id}
                className="bg-slate-800 border border-slate-700 rounded-lg p-6 shadow-lg hover:shadow-xl transition"
              >
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h2 className="text-xl font-bold text-white">
                        {server.name}
                      </h2>
                      <span
                        className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${
                          server.status === "running"
                            ? "bg-green-900/40 text-green-300 border border-green-700"
                            : server.status === "starting"
                              ? "bg-yellow-900/40 text-yellow-300 border border-yellow-700"
                              : "bg-red-900/40 text-red-300 border border-red-700"
                        }`}
                      >
                        {server.status.toUpperCase()}
                      </span>
                    </div>
                    <p className="text-slate-400 text-sm mt-2">
                      Port: {server.port}
                      {server.players !== undefined &&
                        ` • Players: ${server.players}`}
                      {server.ram_mb !== undefined &&
                        ` • RAM: ${server.ram_mb}MB`}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    {server.status === "running" ? (
                      <button
                        onClick={() => stopServer(server.id)}
                        className="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-lg transition"
                      >
                        Stop
                      </button>
                    ) : (
                      <button
                        onClick={() => startServer(server.id)}
                        className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded-lg transition"
                      >
                        Start
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      <footer className="bg-slate-950 border-t border-slate-700 mt-12">
        <div className="max-w-7xl mx-auto px-6 py-4 text-center text-slate-500 text-sm">
          MCAP v0.1.0 • Minecraft Server Management Panel
        </div>
      </footer>
    </div>
  );
}
