'use client'

import { useEffect } from 'react'
import { useServerStore } from '@/store/server'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Server as ServerIcon, Activity, Memory, Cpu } from 'lucide-react'
import Link from 'next/link'

interface ServerStats {
  id: string
  name: string
  status: string
  onlinePlayers: number
  tps: number
  ramMB: number
  cpuPercent: number
}

export default function Dashboard() {
  const { servers, selectServer } = useServerStore()

  const { data: serverList, isLoading } = useQuery<ServerStats[]>({
    queryKey: ['/api/v1/servers'],
    queryFn: async () => {
      const res = await axios.get<ServerStats[]>('/api/v1/servers')
      return res.data
    },
  })

  useEffect(() => {
    if (serverList) {
      // Sync with store
    }
  }, [serverList])

  if (isLoading) {
    return <div className="p-8">Loading...</div>
  }

  return (
    <div className="container mx-auto p-4 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <Button asChild>
          <Link href="/servers/new">Create Server</Link>
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {serverList?.map((server) => (
          <Card key={server.id} className="overflow-hidden transition-all hover:shadow-lg">
            <CardHeader className="bg-gradient-to-r from-gray-800 to-gray-900 text-white">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">{server.name}</CardTitle>
                <span className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                  server.status === 'running' ? 'bg-green-500/20 text-green-400' :
                  server.status === 'stopped' ? 'bg-gray-500/20 text-gray-400' :
                  'bg-red-500/20 text-red-400'
                }`}>
                  {server.status}
                </span>
              </div>
            </CardHeader>
            <CardContent className="space-y-4 p-6">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex items-center space-x-3">
                  <Activity className="h-5 w-5 text-blue-500" />
                  <div>
                    <div className="text-sm text-muted-foreground">TPS</div>
                    <div className="font-semibold">{server.tps.toFixed(1)}</div>
                  </div>
                </div>
                <div className="flex items-center space-x-3">
                  <Memory className="h-5 w-5 text-purple-500" />
                  <div>
                    <div className="text-sm text-muted-foreground">RAM</div>
                    <div className="font-semibold">{server.ramMB} MB</div>
                  </div>
                </div>
              </div>
              <div className="flex items-center space-x-3">
                <Cpu className="h-5 w-5 text-yellow-500" />
                <div>
                  <div className="text-sm text-muted-foreground">CPU</div>
                  <div className="font-semibold">{server.cpuPercent.toFixed(1)}%</div>
                </div>
              </div>
              <div className="pt-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <ServerIcon className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">
                      {server.onlinePlayers} online
                    </span>
                  </div>
                  <div className="flex space-x-2">
                    <Button size="sm" variant="outline">Start</Button>
                    <Button size="sm" variant="outline">Stop</Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
        {(!serverList || serverList.length === 0) && (
          <div className="col-span-full flex flex-col items-center justify-center rounded-lg border border-dashed p-8 text-center">
            <ServerIcon className="h-12 w-12 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">No servers yet</h3>
            <p className="mt-2 text-sm text-muted-foreground">Create your first Minecraft server to get started.</p>
            <Button className="mt-4" asChild>
              <Link href="/servers/new">Create Server</Link>
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
