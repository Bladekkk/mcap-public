# MCAP - Minecraft Server Management Panel

Advanced admin panel for managing Minecraft servers with support for multiple server types (Vanilla, Paper, Fabric, Forge, Quilt, Purpur, Velocity, Waterfall) and automated dependency management.

## Features

- **Auto-installer**: Automatically downloads and caches JDK versions (8/11/17/21) based on Minecraft version via Adoptium API
- **Server management**: Create, start, stop, restart servers with one click
- **Live monitoring**: Real-time TPS, RAM, CPU, online players via WebSocket
- **Plugin/Mod manager**: Search and install plugins from Modrinth, Hangar, CurseForge
- **Java version mapping**: Automatically selects correct Java version for server type
- **Secure**: JWT auth, RCON password protection, audit logging

## Tech Stack

### Backend
- Go 1.22+ with Fiber framework
- SQLite database with migrations
- WebSocket for real-time updates
- Native subprocess control for Minecraft servers

### Frontend
- Next.js 14 with App Router
- React Query for data fetching
- Tailwind CSS + shadcn/ui for UI
- Lucide icons
- Framer Motion for animations

## Project Structure

```
MCAP-project/
├── backend/
│   ├── cmd/
│   │   └── main.go         # Entry point
│   ├── internal/
│   │   ├── config/         # Configuration
│   │   ├── database/       # DB layer
│   │   ├── handler/        # HTTP handlers
│   │   └── service/        # Business logic
│   ├── migrations/         # SQLite migrations
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── app/            # Next.js pages
│   │   ├── components/     # shadcn/ui components
│   │   ├── store/          # Zustand store
│   │   ├── types/          # TypeScript types
│   │   └── lib/            # Utilities
│   └── package.json
├── Dockerfile
└── docker-compose.yml
```

## Installation

### Prerequisites
- Go 1.22+
- Node.js 18+ and npm
- SQLite3 (for migrations)

### Local Development

```bash
# Backend
cd backend
go mod tidy
go run cmd/main.go

# Frontend (new terminal)
cd frontend
npm install
npm run dev
```

### Docker

```bash
docker-compose up --build
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/servers` | List all servers |
| POST | `/api/v1/servers` | Create new server |
| GET | `/api/v1/servers/:id` | Get server details |
| POST | `/api/v1/servers/:id/start` | Start server |
| POST | `/api/v1/servers/:id/stop` | Stop server |
| POST | `/api/v1/servers/:id/restart` | Restart server |
| DELETE | `/api/v1/servers/:id` | Delete server |
| GET | `/api/v1/java` | List installed JDKs |
| POST | `/api/v1/java` | Install JDK |
| GET | `/api/v1/plugins?q=...` | Search plugins |
| GET | `/api/v1/metrics/:id` | Get real-time metrics |
| GET | `/api/v1/metrics/:id/history` | Get metrics history |
| WS | `/ws` | WebSocket for live updates |

## Configuration

Environment variables (create `.env` in `data/` directory):

```
MC_DATA_DIR=/home/user/.mcap
MC_CACHE_DIR=/home/user/.mcap/cache
MC_SERVER_PORT=8080
MC_RCON_PORT=25575
MC_RCON_PASSWORD=change_me_in_production
MC_MODRINTH_TOKEN=your_token
MC_CURSEFORGE_KEY=your_key
```

## Known Limitations

1. **Rate limits**: Modrinth/CurseForge APIs have rate limits - implement caching
2. **Cloudflare**: SpigotMC may block automated requests
3. **Platform**: `taskset` CPU affinity only works on Linux
4. **RCON**: Requires `enable-rcon=true` in server.properties

## TODO

- [ ] Implement file browser for plugins/mods
- [ ] Add automated backups with scheduling
- [ ] Implement WebSocket for console input/output
- [ ] Add alerting (Telegram/Discord webhooks)
- [ ] Support for Velocity/Waterfall proxy servers
- [ ] Implement plugin dependency resolution

## License

MIT
