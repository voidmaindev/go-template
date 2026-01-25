# API Visual Tester Frontend

A cyberpunk-themed visual testing interface for the Go Template backend API.

## Features

- **Authentication**: Login, Register, Logout, Token refresh
- **User Management**: View, delete users, manage roles
- **Items CRUD**: Full create, read, update, delete operations
- **Countries CRUD**: Manage countries with ISO codes
- **Cities CRUD**: Manage cities linked to countries
- **Documents**: Create documents with line items, manage items
- **RBAC Management**: Create roles, manage permissions

## Tech Stack

- **React 18** with TypeScript
- **Vite** for build tooling
- **TailwindCSS** for styling
- **React Query** for server state
- **React Router** for navigation
- **Zustand** for client state
- **React Hook Form** for forms
- **Framer Motion** for animations
- **Lucide React** for icons

## Quick Start

### Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at http://localhost:5178

### Docker

```bash
# Production build
docker compose up -d

# Development with hot reload
docker compose --profile dev up frontend-dev
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```
VITE_API_URL=http://localhost:3000/api/v1
```

## Project Structure

```
frontend/
├── src/
│   ├── components/     # Reusable UI components
│   ├── hooks/          # Custom React hooks (API calls)
│   ├── lib/            # Utilities and API client
│   ├── pages/          # Page components
│   ├── store/          # Zustand stores
│   ├── types/          # TypeScript types
│   ├── App.tsx         # Root component
│   └── main.tsx        # Entry point
├── public/             # Static assets
├── Dockerfile          # Production Docker image
├── docker-compose.yml  # Docker Compose config
└── package.json        # Dependencies
```

## API Endpoints Covered

| Domain | Endpoints |
|--------|-----------|
| Auth | login, register, refresh, logout |
| Users | list, get, delete, profile, password |
| Items | list, get, create, update, delete |
| Countries | list, get, create, update, delete, cities |
| Cities | list, get, create, update, delete |
| Documents | list, get, create, update, delete, items |
| RBAC | roles, permissions, user roles, domains |

## Design

The interface uses a "Terminal Hacker" aesthetic with:
- Dark cyber theme (#0a0a0f base)
- Neon accent colors (cyan, pink, green)
- Orbitron display font
- JetBrains Mono monospace font
- Rajdhani body font
- Scan-line effects
- Glow effects on interactive elements
