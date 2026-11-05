export interface Server {
  id: string
  name: string
  jarPath: string
  status: 'stopped' | 'starting' | 'running' | 'stopping' | 'crashed'
  port: number
  javaVersion: string
  xmsMB: number
  xmxMB: number
  autostart: boolean
  eulaAccepted: boolean
  createdAt: string
  onlinePlayers?: number
  tps?: number
  ramMB?: number
  cpuPercent?: number
}

export interface Plugin {
  id: string
  name: string
  description: string
  author: string
  downloads: number
  version: string
  source: 'modrinth' | 'hangar' | 'curseforge'
  downloadUrl?: string
}

export interface MetricsHistory {
  timestamp: string
  tps: number
  mspt: number
  onlinePlayers: number
  ramMB: number
  cpuPercent: number
}
