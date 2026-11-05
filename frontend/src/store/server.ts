import { create } from 'zustand'
import { Server } from '@/types'

interface ServerState {
  servers: Server[]
  selectedServerId: string | null
  addServer: (server: Server) => void
  updateServer: (id: string, updates: Partial<Server>) => void
  removeServer: (id: string) => void
  selectServer: (id: string | null) => void
}

export const useServerStore = create<ServerState>((set) => ({
  servers: [],
  selectedServerId: null,
  addServer: (server) => set((state) => ({ servers: [...state.servers, server] })),
  updateServer: (id, updates) => set((state) => ({
    servers: state.servers.map((s) => (s.id === id ? { ...s, ...updates } : s)),
  })),
  removeServer: (id) => set((state) => ({
    servers: state.servers.filter((s) => s.id !== id),
    selectedServerId: state.selectedServerId === id ? null : state.selectedServerId,
  })),
  selectServer: (id) => set({ selectedServerId: id }),
}))
