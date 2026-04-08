# My PaaS - Docker Visual Designer

> Visual drag-and-drop Docker service designer for VS Code with auto Dockerfile generation and docker-compose management.

## Overview

My PaaS is a VS Code extension that lets you visually design multi-container Docker applications on an interactive canvas, then generate and manage `docker-compose.yml` — no YAML editing required.

**Key Techniques for Graceful Deployments & Zero Downtime Deploy:**
- `stop_grace_period`: Define how long Docker waits before killing a container after sending a SIGTERM.
- `Healthchecks`: Ensure new containers are fully running and passing checks before the old ones stop.
- `docker-rollout` Plugin: Use the wowu/docker-rollout tool for zero-downtime updates with native Docker Compose commands.
- `Blue-Green Deployment`: Run two sets of services and use a reverse proxy (like Caddy or Nginx) to toggle traffic between them, ensuring the old version runs until the new one is live

```
my-paas/
├── core/                  # Go CLI — auto-detect & Dockerfile generation
│   ├── cmd/               # CLI entry point (detect, dockerfile)
│   ├── app/               # Source directory scanner
│   ├── provider/          # Language/framework providers
│   ├── generate/          # Dockerfile generator
│   └── plan/              # Build plan model
├── my-paas-extension/     # VS Code extension
│   ├── src/               # Extension TypeScript source
│   │   ├── core/          # CoreBridge — Go binary integration
│   │   ├── docker/        # DockerManager, ComposeGenerator
│   │   ├── panels/        # Canvas webview panel
│   │   ├── sidebar/       # Service & template tree views
│   │   └── templates/     # Built-in service templates
│   └── webview/           # React + React Flow canvas UI
└── docs/                  # Design docs
```

## Features

- **Visual Canvas** — drag-and-drop service designer powered by React Flow
- **Auto-Detect** — automatically detect language, framework, and generate Dockerfile from source code
- **Docker Compose** — generate `docker-compose.yml` from canvas layout (services, links, volumes, ports, env vars)
- **Compose Controls** — run `docker compose up/down` directly from VS Code
- **14 Built-in Templates** — PostgreSQL, MySQL, MongoDB, Redis, Traefik, Nginx, RabbitMQ, MinIO, Node.js, Go, Python, and more
- **Project Persistence** — canvas state saved per workspace

### Supported Languages (Auto-Detect)

| Language | Frameworks |
|----------|-----------|
| Node.js  | React, Next.js, Nuxt, Svelte, Astro, Remix, etc. |
| Go       | Standard, multi-module workspaces |
| Python   | Django, Flask, FastAPI |
| Java     | Maven, Gradle |
| PHP      | Laravel, Symfony |
| Rust     | Cargo |
| Static   | HTML/CSS/JS |

## Prerequisites

- [VS Code](https://code.visualstudio.com/) 1.85+
- [Go](https://go.dev/) 1.26+
- [Bun](https://bun.sh/)
- [Docker](https://www.docker.com/) (for compose up/down)

## Getting Started

```bash
# Clone
git clone https://github.com/my-paas/my-paas.git
cd my-paas

# Install dependencies
cd my-paas-extension
bun run install:all

# Build everything (Go core + extension + webview)
bun run build

# Development — watch mode
bun run watch          # Extension TypeScript
bun run dev:webview    # Webview React (hot reload)
```

Press `F5` in VS Code to launch the Extension Development Host.

## Commands

| Command | Description |
|---------|-------------|
| `My PaaS: Open Canvas` | Open the visual service designer |
| `My PaaS: Auto Detect App` | Detect language/framework from source |
| `My PaaS: Generate docker-compose.yml` | Generate compose file from canvas |
| `My PaaS: Compose Up` | Start all services |
| `My PaaS: Compose Down` | Stop all services |

## Packaging

```bash
cd my-paas-extension

# Package for current platform
bun run package

# Package for a specific platform
bun run package:win-x64
bun run package:linux-x64
bun run package:darwin-arm64

# Package all 6 platforms
bun run package:all

# Publish to VS Code Marketplace
bun run publish
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Extension | TypeScript, VS Code Extension API |
| Core CLI | Go 1.26 |
| Webview | React 18, React Flow, Zustand, Vite |
| Build | Bun, esbuild |

## License

MIT
