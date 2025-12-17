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
- **Language**: Go 1.25+
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
   git clone github.com/vitaly-stepin/agile_party
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


## License

MIT
