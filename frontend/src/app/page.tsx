'use client'

import { useState, useEffect } from 'react'
import axios from 'axios'
import Link from 'next/link'

interface Server {
  id: string
  name: string
  status: string
  jar_path: string
  java_version: string
  xms_mb: number
  xmx_mb: number
  port: number
}

export default function Dashboard() {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchServers = async () => {
      try {
        const res = await axios.get('/api/v1/servers')
        setServers(res.data.servers)
      } catch (err) {
        console.error('Failed to fetch servers:', err)
      } finally {
        setLoading(false)
      }
    }
    fetchServers()
  }, [])

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <div className="container mx-auto p-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-4xl font-bold text-blue-400">MCAP Dashboard</h1>
          <Link
            href="/servers/new"
            className="px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition"
          >
            Create Server
          </Link>
        </div>

        {loading ? (
          <div className="text-center py-12">Loading servers...</div>
        ) : servers.length === 0 ? (
          <div className="text-center py-12 border-2 border-dashed border-gray-700 rounded-lg">
            <h2 className="text-2xl font-semibold mb-4">No servers yet</h2>
            <Link href="/servers/new" className="text-blue-400 hover:underline">
              Create your first server
            </Link>
          </div>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {servers.map((server) => (
              <div
                key={server.id}
                className="bg-gray-800 rounded-lg p-6 shadow-lg hover:shadow-xl transition"
              >
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-xl font-semibold">{server.name}</h2>
                  <span
                    className={`px-2 py-1 rounded text-sm font-medium ${
                      server.status === 'running'
                        ? 'bg-green-600 text-white'
                        : 'bg-gray-600 text-white'
                    }`}
                  >
                    {server.status}
                  </span>
                </div>
                <div className="space-y-2 text-gray-300 text-sm">
                  <p>
                    <span className="font-medium">Jar:</span> {server.jar_path}
                  </p>
                  <p>
                    <span className="font-medium">Java:</span> {server.java_version}
                  </p>
                  <p>
                    <span className="font-medium">Memory:</span> {server.xms_mb}MB / {server.xmx_mb}MB
                  </p>
                  <p>
                    <span className="font-medium">Port:</span> {server.port}
                  </p>
                </div>
                <div className="mt-6 flex space-x-2">
                  <button className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 rounded transition">
                    Start
                  </button>
                  <button className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 rounded transition">
                    Stop
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
