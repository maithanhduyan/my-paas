# Tổng quan kiến trúc

## Mô hình tổng thể

My PaaS sử dụng kiến trúc **monolith modular** — một Go binary duy nhất đảm nhiệm toàn bộ logic, giao tiếp với Docker Engine qua socket mount.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                             │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌────────────┐  │
│  │ Web UI    │  │ CLI Tool  │  │ VS Code   │  │ API Client │  │
│  │ (React)   │  │ (mypaas)  │  │ Extension │  │ (curl, etc)│  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬──────┘  │
└────────┼───────────────┼──────────────┼──────────────┼──────────┘
         │               │              │              │
         └───────────────┼──────────────┼──────────────┘
                         │              │
                    HTTPS/HTTP     Docker compose
                         │         (local only)
┌────────────────────────▼──────────────────────────────────────┐
│                    Traefik (Reverse Proxy)                     │
│              Port 80/443 · Auto-SSL · Routing                 │
└────────────────────────┬──────────────────────────────────────┘
                         │
┌────────────────────────▼──────────────────────────────────────┐
│                   My PaaS Server (:8080)                       │
│                                                                │
│  ┌─ Middleware Stack ──────────────────────────────────────┐   │
│  │ Recover → RequestID → SecurityHeaders → Logger →        │   │
│  │ CORS → Metrics → RateLimiter → [AuthRequired]           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                │
│  ┌─ Handler Layer (handler/*.go) ──────────────────────────┐   │
│  │ auth · project · deployment · service · domain · env    │   │
│  │ logs · stats · backup · marketplace · samples · swarm   │   │
│  │ enterprise · registry · openapi · health · webhook      │   │
│  └───────────┬──────────────┬──────────────┬───────────────┘   │
│              │              │              │                    │
│  ┌───────────▼──┐  ┌───────▼──────┐  ┌───▼──────────────┐     │
│  │ Store Layer  │  │ Docker Client│  │ Worker Queue     │     │
│  │ SQLite/PG    │  │ Docker SDK   │  │ (N concurrent)   │     │
│  └───────┬──────┘  └──────┬───────┘  └──────┬───────────┘     │
│          │                │                  │                 │
│  ┌───────▼──┐      ┌──────▼───────┐   ┌─────▼────────────┐    │
│  │ Redis    │      │ Git Watcher  │   │ Notifier         │    │
│  │(optional)│      │ (60s poll)   │   │(webhook/slack)   │    │
│  └──────────┘      └──────────────┘   └──────────────────┘    │
└────────────────────────┬──────────────────────────────────────┘
                         │ Docker Socket
                         │ /var/run/docker.sock
┌────────────────────────▼──────────────────────────────────────┐
│                   Docker Engine / Swarm                        │
│                                                                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────────────────┐  │
│  │ App #1  │ │ App #2  │ │ App #N  │ │ Backing Services   │  │
│  │ (node)  │ │ (python)│ │ (go)    │ │ PG · Redis · Mongo │  │
│  └─────────┘ └─────────┘ └─────────┘ └────────────────────┘  │
│                                                                │
│  ┌─────────────────────┐  ┌───────────────────────────────┐   │
│  │ Local Registry      │  │ mypaas-network (overlay)      │   │
│  │ (registry:2)        │  │ Service discovery · DNS       │   │
│  └─────────────────────┘  └───────────────────────────────┘   │
└───────────────────────────────────────────────────────────────┘
```

## Các thành phần chính

### 1. HTTP Server (Fiber v2)

Framework Go hiệu suất cao dựa trên fasthttp. Xử lý toàn bộ REST API + phục vụ static files (React SPA).

**Middleware stack** thực thi theo thứ tự:
1. **Recover** — bắt panic, trả 500 thay vì crash
2. **RequestID** — gán UUID cho mỗi request để trace
3. **SecurityHeaders** — HSTS, CSP, X-Frame-Options, X-Content-Type-Options
4. **Logger** — format: `{time} {status} {method} {path} {latency}`
5. **CORS** — cho phép cross-origin từ web dashboard
6. **Metrics** — đếm request, latency per route
7. **RateLimiter** — Redis-backed hoặc in-memory sliding window
8. **AuthRequired** — JWT → API Key → Session Token (triple auth)

### 2. Store Layer

Abstraction trên database với dual-driver support:

- **SQLite** (default) — WAL mode, `?` placeholders
- **PostgreSQL** — `$1` placeholders, connection pooling

18 bảng dữ liệu, migration tự động khi khởi động (`AutoMigrate()`).

### 3. Docker Client

Wrap Docker SDK (`github.com/docker/docker`), cung cấp hai chế độ:

- **Container mode** — `docker run/stop/rm` trực tiếp
- **Swarm mode** — `docker service create/update/rm` với replicas, rolling update

Phát hiện mode tự động: nếu `docker swarm` active → dùng Swarm API.

### 4. Deploy Workers

Go channel-based job queue:

```
Request → Queue (buffered channel, size=100)
                  ↓
         Worker Pool (N goroutines)
                  ↓
         DeployWorker.Handle()
           1. Clone repo
           2. Detect language
           3. Generate Dockerfile
           4. Build image
           5. Deploy container/service
           6. Health check
```

Mỗi bước ghi log real-time vào `deployment_logs`, client nhận qua SSE.

### 5. Git Watcher

Background goroutine poll mỗi 60 giây:
- Lấy tất cả project có `auto_deploy = true`
- So sánh `git ls-remote HEAD` với commit hash lưu trong DB
- Nếu khác → queue deploy job

### 6. Core Module

Module Go riêng biệt (`core/`) — detect ngôn ngữ và sinh Dockerfile:

| Provider | Detect bằng | Framework detection |
|---|---|---|
| PHP | `composer.json` | Laravel, Symfony |
| Go | `go.mod` | Gin, Fiber, Echo |
| Java | `pom.xml`, `build.gradle` | Spring Boot |
| Rust | `Cargo.toml` | Actix, Rocket |
| Python | `requirements.txt`, `Pipfile`, `pyproject.toml` | Django, Flask, FastAPI |
| Node.js | `package.json` | Next.js, React, Express, Vite |
| Static | `index.html` | HTML/CSS |

Output: `BuildPlan` → `GenerateDockerfile()` → multi-stage Dockerfile tối ưu.

## Luồng dữ liệu

### Deploy flow

```
Git Push → Webhook/Watcher/Manual
    │
    ▼
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Clone   │───►│  Detect  │───►│  Build   │───►│  Deploy  │
│ git clone│    │core.Plan │    │docker    │    │swarm svc │
│ --depth 1│    │          │    │build     │    │or docker │
└──────────┘    └──────────┘    └──────────┘    │run       │
                                                └────┬─────┘
                                                     │
                                                     ▼
                                               ┌──────────┐
                                               │  Health  │
                                               │  Check   │
                                               │(60/90s)  │
                                               └────┬─────┘
                                                     │
                                            ┌────────┴────────┐
                                            ▼                 ▼
                                       ✅ healthy        ❌ failed
                                       (stop old)       (rollback)
```

### Auth flow

```
Request
  │
  ├─ Header: "Bearer eyJ..."  → JWT decode → user_id, role
  │
  ├─ Header: "Bearer mpk_..."  → SHA-256 hash → lookup api_keys
  │
  └─ Header: "Bearer sess_..." → lookup sessions table → user_id
```

## Scaling

| Quy mô | Cấu hình | Ước lượng |
|---|---|---|
| **1 developer** | 1 node, SQLite, no Redis | 10–20 apps, $5/tháng |
| **Team nhỏ** | 2 nodes Swarm, SQLite | 20–50 apps, $15/tháng |
| **Startup** | 2+ nodes, PostgreSQL, Redis | 50–100 apps, $30/tháng |
| **Enterprise** | 3+ nodes, PG + Redis, monitoring | 100+ apps |
