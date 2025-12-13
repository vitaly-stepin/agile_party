# Agile Party - A Scrum Poker tool to plan faster and have more fun doing it 🚀

A real-time Scrum Poker estimation tool built with Go (backend) and React (frontend).

## Project Structure

```
agile_party/
├── backend/          # Go backend with Hexagonal Architecture
│   ├── cmd/api/      # Application entry point
│   ├── internal/     # Private application code
│   │   ├── domain/   # Core business logic
│   │   ├── application/  # Use cases
│   │   ├── adapters/     # External system adapters
│   │   └── interfaces/   # HTTP/WebSocket handlers
│   └── pkg/          # Public utilities
├── frontend/         # React + TypeScript + Tailwind CSS
└── docker-compose.yml
```

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Fiber (HTTP + WebSocket)
- **Database**: PostgreSQL 16
- **Architecture**: Hexagonal (Ports & Adapters)

### Frontend
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Routing**: React Router

## Quick Start

### Prerequisites
- Docker & Docker Compose

### Running Locally

1. **Clone the repository**
   ```bash
   git clone <repo-url>
   cd agile_party
   ```

2. **Copy environment variables**
   ```bash
   cp .env.example .env
   ```

3. **Start all services**
   ```bash
   docker-compose up
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/api/health

### Development

The setup includes hot reload for both backend and frontend:
- **Backend**: Air watches Go files and rebuilds automatically
- **Frontend**: Vite dev server with Hot Module Replacement (HMR)

## Phase 1 Status ✅ COMPLETED (2025-12-14)

- [x] Backend Go module initialized (`github.com/vitaly-stepin/agile_party`)
- [x] Backend directory structure created (Hexagonal Architecture)
- [x] Frontend with Vite + React 18 + TypeScript
- [x] Tailwind CSS v4 configured with `@tailwindcss/postcss`
- [x] Docker Compose setup (PostgreSQL 16 + Backend + Frontend)
- [x] Air hot reload configured (`github.com/air-verse/air@latest`)
- [x] Environment configuration (.env.example)
- [x] Basic health check endpoint (`GET /api/health`)
- [x] All services running successfully

**Verified Working:**
- Backend: http://localhost:8080/api/health → `{"status":"ok"}`
- Frontend: http://localhost:5173 → Vite + React running
- PostgreSQL: localhost:5432 → Database ready

## Next Steps (Phase 2: Domain Layer)

**To continue in a new session, simply say:**
> "Let's implement Phase 2 (Domain Layer)"

The plan file contains all context: `/Users/vitaliy-stepin/.claude/plans/clever-doodling-scroll.md`

**What Phase 2 will implement:**
1. Room entity with validation
2. User entity
3. Vote value object
4. Domain services (voting validation, average calculation)
5. Repository interfaces (ports)
6. Comprehensive unit tests

## License

MIT
