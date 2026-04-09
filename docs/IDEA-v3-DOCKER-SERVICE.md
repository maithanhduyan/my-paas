# My PaaS v3 - Docker Service quản lý CI/CD

> Một Docker service duy nhất chạy trên host, quản lý toàn bộ deploy/CI/CD cho các dự án khác — giống Railway self-hosted.

---

## 1. Tổng quan kiến trúc

```
┌─────────────────────────────────────────────────────────────────┐
│                        HOST MACHINE                             │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              MY-PAAS CONTAINER (Docker Service)           │  │
│  │                                                           │  │
│  │  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │  │
│  │  │ API     │  │ Worker   │  │ Watcher  │  │ Frontend │  │  │
│  │  │ Server  │  │ (Build/  │  │ (Git +   │  │ UI       │  │  │
│  │  │ (Go)    │  │  Deploy) │  │  .env)   │  │ (React)  │  │  │
│  │  └────┬────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │  │
│  │       │            │             │              │         │  │
│  │  ┌────┴────────────┴─────────────┴──────────────┴─────┐  │  │
│  │  │                 Shared Components                   │  │  │
│  │  │  ┌────────┐ ┌─────────┐ ┌────────┐ ┌───────────┐  │  │  │
│  │  │  │SQLite/ │ │ Redis/  │ │ Core   │ │ Docker    │  │  │  │
│  │  │  │Postgres│ │ Queue   │ │ Detect │ │ Engine    │  │  │  │
│  │  │  │        │ │         │ │        │ │ Client    │  │  │  │
│  │  │  └────────┘ └─────────┘ └────────┘ └─────┬─────┘  │  │  │
│  │  └───────────────────────────────────────────┼────────┘  │  │
│  └──────────────────────────────────────────────┼────────────┘  │
│                                                 │               │
│  ┌──────────────────────────────────────────────┼────────────┐  │
│  │                Docker Engine (Host)          │            │  │
│  │                                              ▼            │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │  │
│  │  │ App 1    │ │ App 2    │ │ Postgres │ │ Redis    │    │  │
│  │  │ (Next.js)│ │ (Go API) │ │          │ │          │    │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘    │  │
│  │                                                           │  │
│  │  ┌──────────┐                                             │  │
│  │  │ Traefik  │ ← Reverse Proxy + Auto SSL                 │  │
│  │  └──────────┘                                             │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**Ý tưởng cốt lõi:** My PaaS chạy như **1 Docker container** (hoặc compose stack), mount Docker socket của host (`/var/run/docker.sock`) để quản lý tất cả container khác.

---

## 2. Các thành phần chính

### 2.1 API Server (Go - Fiber/Echo)

Tận dụng lại `core/` module hiện tại cho auto-detect + Dockerfile generation.

```
api/
├── main.go
├── handler/
│   ├── project.go        # CRUD dự án
│   ├── deployment.go     # Trigger/cancel/rollback deploy
│   ├── service.go        # Quản lý database services (Postgres, Redis...)
│   ├── environment.go    # Quản lý env vars
│   ├── domain.go         # Custom domain + SSL
│   ├── logs.go           # Stream logs (SSE/WebSocket)
│   └── webhook.go        # GitHub/GitLab webhook receiver
├── model/
│   ├── project.go
│   ├── deployment.go
│   ├── service.go
│   └── environment.go
├── worker/
│   ├── builder.go        # Build Docker image từ source
│   ├── deployer.go       # Start/stop/replace container
│   └── watcher.go        # Watch git changes + env changes
├── docker/
│   ├── client.go         # Docker Engine API wrapper
│   ├── network.go        # Quản lý Docker networks
│   ├── volume.go         # Quản lý Docker volumes
│   └── registry.go       # Local registry (optional)
├── proxy/
│   └── traefik.go        # Dynamic Traefik config generation
└── store/
    ├── sqlite.go          # SQLite cho simplicity (hoặc Postgres)
    └── migrations/
```

### 2.2 Worker System

Job queue để xử lý build/deploy async:

```go
// Job types
const (
    JobCloneRepo    = "clone_repo"
    JobDetectApp    = "detect_app"
    JobBuildImage   = "build_image"
    JobDeployApp    = "deploy_app"
    JobRollback     = "rollback"
    JobCleanup      = "cleanup"
)
```

**Pipeline deploy:**
```
GitHub Push → Webhook → Queue Job
                          │
                          ▼
              ┌─── clone_repo ───┐
              │                  │
              ▼                  │
         detect_app              │
              │                  │
              ▼                  │
         build_image             │ Mỗi step log
              │                  │ realtime
              ▼                  │
         deploy_app              │
              │                  │
              ▼                  │
         health_check ───────────┘
              │
              ├── OK → Mark "healthy" + update proxy
              └── FAIL → Auto rollback to previous version
```

### 2.3 Watcher System

Giám sát thay đổi và trigger deploy tự động:

```go
type Watcher struct {
    projects map[string]*ProjectWatch
}

type ProjectWatch struct {
    ProjectID   string
    GitRepo     string
    Branch      string
    LastCommit  string
    EnvHash     string        // Hash của .env hiện tại
    PollInterval time.Duration // Mặc định 30s
    Status      WatchStatus   // watching | ready_to_deploy | deploying
}

type WatchStatus string
const (
    StatusWatching     WatchStatus = "watching"
    StatusReadyDeploy  WatchStatus = "ready_to_deploy"   // .env changed
    StatusDeploying    WatchStatus = "deploying"
    StatusHealthy      WatchStatus = "healthy"
    StatusFailed       WatchStatus = "failed"
)
```

**Hai chế độ trigger:**

| Trigger | Hành vi |
|---------|---------|
| **Git push** (webhook/poll) | Tự động deploy ngay |
| **.env thay đổi** (UI edit) | Chuyển sang `ready_to_deploy`, chờ user confirm |

**Cách hoạt động:**

1. **Webhook mode** (ưu tiên): GitHub/GitLab gửi webhook khi có push → deploy ngay
2. **Polling mode** (fallback): Poll git remote mỗi 30s, so sánh commit hash
3. **Env change**: User sửa env qua UI → hash thay đổi → badge "Ready to Deploy" → user click Deploy

### 2.4 Frontend UI (React + Vite)

```
webview/src/
├── App.tsx
├── pages/
│   ├── Dashboard.tsx         # Overview tất cả projects
│   ├── ProjectDetail.tsx     # Chi tiết 1 project
│   ├── Deployments.tsx       # Lịch sử deploy
│   ├── Services.tsx          # Database/Cache services
│   ├── Settings.tsx          # Domain, env vars, scaling
│   └── Canvas.tsx            # Drag-drop service designer (reuse v2)
├── components/
│   ├── ServiceCard.tsx       # Card hiển thị 1 service
│   ├── DeployButton.tsx      # Deploy / Ready to Deploy badge
│   ├── LogViewer.tsx         # Realtime log stream
│   ├── EnvEditor.tsx         # Editor cho .env variables
│   ├── DomainConfig.tsx      # Custom domain setup
│   ├── MetricsChart.tsx      # CPU/RAM charts
│   └── Canvas/               # Drag-drop canvas components (từ v2)
│       ├── CanvasBoard.tsx
│       ├── ServiceNode.tsx
│       └── ConnectionEdge.tsx
└── api/
    └── client.ts             # API client (fetch/axios)
```

---

## 3. Database Schema

```sql
-- Dự án / Application
CREATE TABLE projects (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    git_url     TEXT,              -- https://github.com/user/repo.git
    branch      TEXT DEFAULT 'main',
    provider    TEXT,              -- auto-detected: node, go, python...
    framework   TEXT,              -- nextjs, gin, django...
    auto_deploy BOOLEAN DEFAULT true,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Environment variables
CREATE TABLE environments (
    id         TEXT PRIMARY KEY,
    project_id TEXT REFERENCES projects(id),
    key        TEXT NOT NULL,
    value      TEXT NOT NULL,      -- encrypted at rest
    is_secret  BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, key)
);

-- Deploy history
CREATE TABLE deployments (
    id          TEXT PRIMARY KEY,
    project_id  TEXT REFERENCES projects(id),
    commit_hash TEXT,
    commit_msg  TEXT,
    status      TEXT DEFAULT 'queued',  -- queued|cloning|building|deploying|healthy|failed|rolled_back
    image_tag   TEXT,                   -- docker image tag
    trigger     TEXT,                   -- webhook|manual|env_change|rollback
    started_at  DATETIME,
    finished_at DATETIME,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Deploy logs (per step)
CREATE TABLE deployment_logs (
    id            TEXT PRIMARY KEY,
    deployment_id TEXT REFERENCES deployments(id),
    step          TEXT,           -- clone|detect|build|deploy|healthcheck
    level         TEXT,           -- info|warn|error
    message       TEXT,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Managed services (databases, caches)
CREATE TABLE services (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL,       -- postgres|redis|mysql|mongo|minio
    image      TEXT NOT NULL,       -- postgres:16-alpine
    status     TEXT DEFAULT 'stopped',
    config     TEXT,                -- JSON: ports, volumes, env
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Service links (project <-> service)
CREATE TABLE service_links (
    id         TEXT PRIMARY KEY,
    project_id TEXT REFERENCES projects(id),
    service_id TEXT REFERENCES services(id),
    env_prefix TEXT,              -- e.g. DATABASE_ → DATABASE_URL, DATABASE_HOST
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, service_id)
);

-- Custom domains
CREATE TABLE domains (
    id         TEXT PRIMARY KEY,
    project_id TEXT REFERENCES projects(id),
    domain     TEXT UNIQUE NOT NULL,
    ssl_auto   BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 4. API Endpoints

```
# Projects
GET    /api/projects                    # List all projects
POST   /api/projects                    # Create project (từ git URL hoặc local path)
GET    /api/projects/:id                # Get project detail
PUT    /api/projects/:id                # Update project settings
DELETE /api/projects/:id                # Delete project + containers

# Deployments
POST   /api/projects/:id/deploy        # Trigger manual deploy
GET    /api/projects/:id/deployments    # List deployment history
GET    /api/deployments/:id             # Deploy detail + status
POST   /api/deployments/:id/rollback   # Rollback to this version
POST   /api/deployments/:id/cancel     # Cancel in-progress deploy

# Environment
GET    /api/projects/:id/env           # List env vars
PUT    /api/projects/:id/env           # Bulk update env vars → status "ready_to_deploy"
DELETE /api/projects/:id/env/:key      # Delete env var

# Logs
GET    /api/projects/:id/logs          # Container logs (SSE stream)
GET    /api/deployments/:id/logs       # Build/deploy logs (SSE stream)

# Services (databases, caches)
GET    /api/services                   # List managed services
POST   /api/services                   # Create service (postgres, redis...)
DELETE /api/services/:id               # Delete service
POST   /api/services/:id/start        # Start service
POST   /api/services/:id/stop         # Stop service
POST   /api/services/:id/link/:projectId  # Link service to project

# Domains
POST   /api/projects/:id/domains      # Add custom domain
DELETE /api/domains/:id                # Remove domain

# Webhooks
POST   /api/webhooks/github           # GitHub push webhook
POST   /api/webhooks/gitlab           # GitLab push webhook

# System
GET    /api/health                     # Health check
GET    /api/stats                      # System stats (CPU, RAM, disk)
```

---

## 5. Deploy Flow chi tiết

### 5.1 Tạo project mới

```
User → UI: "New Project" → Nhập Git URL
                              │
                              ▼
                    API: POST /api/projects
                              │
                              ▼
                    Clone repo → core.Detect()
                              │
                              ▼
                    Hiển thị: "Node.js / Next.js detected"
                    Hiển thị: suggested env vars
                              │
                              ▼
                    User confirm → First Deploy
```

### 5.2 Auto Deploy (git push)

```
Developer push code → GitHub Webhook
                         │
                         ▼
              POST /api/webhooks/github
                         │
                         ▼
              Verify signature (HMAC)
                         │
                         ▼
              Match project by repo URL
                         │
                         ▼
              Create deployment record (status: queued)
                         │
                         ▼
              Queue: clone → detect → build → deploy → healthcheck
                         │
                         ▼
              Stream logs to UI via SSE
                         │
                         ├── OK → Mark healthy, update Traefik
                         └── FAIL → Rollback, notify user
```

### 5.3 Env Change → Ready to Deploy

```
User sửa env vars qua UI
        │
        ▼
PUT /api/projects/:id/env
        │
        ▼
Update DB + Hash env mới
        │
        ▼
Project status → "ready_to_deploy"
        │
        ▼
UI hiển thị badge:  🟡 Ready to Deploy
        │
        ▼
User click "Deploy" → re-deploy với env mới
(KHÔNG auto deploy khi env thay đổi — tránh downtime bất ngờ)
```

---

## 6. Docker Management

### 6.1 Mount Docker Socket

```yaml
# docker-compose.yml để chạy My PaaS
version: "3.8"
services:
  mypaas:
    image: mypaas:latest
    ports:
      - "3000:3000"     # UI
      - "8080:8080"     # API
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # ← KEY: quản lý Docker host
      - mypaas-data:/data                          # SQLite + project cache
      - mypaas-builds:/builds                      # Build cache
    environment:
      - MYPAAS_SECRET=your-secret-key
      - MYPAAS_DOMAIN=mypaas.local
    networks:
      - mypaas-network

  traefik:
    image: traefik:v3
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - traefik-certs:/certs
    command:
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.email=you@email.com
      - --certificatesresolvers.letsencrypt.acme.storage=/certs/acme.json
      - --certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web
    networks:
      - mypaas-network

volumes:
  mypaas-data:
  mypaas-builds:
  traefik-certs:

networks:
  mypaas-network:
    driver: bridge
```

### 6.2 Container Naming Convention

```
mypaas-{project_name}-{deployment_id}
```

Ví dụ:
- `mypaas-my-blog-d7f3a2` (app container)
- `mypaas-svc-postgres-main` (managed service)
- `mypaas-svc-redis-cache` (managed service)

### 6.3 Blue-Green Deploy

```
Deployment v1 running: mypaas-blog-abc123
                          │
New deploy triggered       │
                          ▼
Build new image: mypaas-blog:def456
                          │
Start new container: mypaas-blog-def456
                          │
Health check (HTTP GET /)  │
                          │
├── Healthy → Update Traefik labels → Route traffic to new
│              Stop old container (mypaas-blog-abc123)
│              Remove old container after 5 min
│
└── Unhealthy → Remove new container
                Keep old running
                Mark deployment as "failed"
```

---

## 7. Networking & Proxy

### Traefik Dynamic Config

Mỗi khi deploy thành công, My PaaS tạo container với Docker labels:

```go
labels := map[string]string{
    "traefik.enable": "true",
    "traefik.http.routers." + name + ".rule":
        "Host(`" + subdomain + "." + baseDomain + "`)",
    "traefik.http.routers." + name + ".entrypoints": "websecure",
    "traefik.http.routers." + name + ".tls.certresolver": "letsencrypt",
    "traefik.http.services." + name + ".loadbalancer.server.port":
        strconv.Itoa(port),
}
```

**Kết quả:**
- `my-blog.mypaas.dev` → container `mypaas-blog-abc123:3000`
- `api.mypaas.dev` → container `mypaas-api-xyz789:8080`

---

## 8. Tận dụng code hiện tại

### Từ `core/`
- **`core.Detect(path)`** → Auto-detect language/framework cho project mới
- **`core.GenerateDockerfile(path)`** → Generate Dockerfile khi build
- **Provider system** → Extend thêm providers nếu cần

### Từ `my-paas-extension/` (webview)
- **Canvas components** → Reuse cho drag-drop service designer trong web UI
- **Template system** → Service templates (Postgres, Redis, MongoDB...)
- **ComposeGenerator** → Reference cho compose generation logic

### Từ `railpack/`
- **Build system** → Reference cho optimized Docker builds (multi-stage, caching)
- **Config merging** → Pattern cho project config (CLI > Env > File)
- **Provider architecture** → Proven pattern cho language detection

---

## 9. Phát triển theo Phase

### Phase 1: Core Engine ✅ DONE & TESTED (2026-04-08)

```
✅ API Server (Go Fiber) — 30 handlers, Fiber v2.52.6
✅ SQLite database + migrations — auto-create tables + indexes
✅ Project CRUD — POST/GET/PUT/DELETE tested
✅ core.Detect() + core.GenerateDockerfile() integration
✅ Docker client (build image, run container, stop, remove)
✅ Manual deploy (git clone → detect → build → deploy pipeline)
✅ Container logs streaming (SSE) — deployment logs endpoint
✅ Basic env vars management — upsert, mask secrets, delete
✅ Worker queue system — 2 goroutine workers, async deploy
✅ Health check endpoint — Docker ping + Go version
```

**Test Results (2026-04-08):**
| API Endpoint | Method | Status |
|---|---|---|
| `/api/health` | GET | ✅ `{"status":"ok","docker":"connected","go":"go1.26.1"}` |
| `/api/projects` | POST | ✅ Create project with git_url |
| `/api/projects` | GET | ✅ List all projects |
| `/api/projects/:id` | GET | ✅ Get single project |
| `/api/projects/:id` | PUT | ✅ Update project name |
| `/api/projects/:id` | DELETE | ✅ Cascading delete |
| `/api/projects/:id/env` | PUT | ✅ Bulk set env vars, status→ready_to_deploy |
| `/api/projects/:id/env` | GET | ✅ List env vars, secrets masked as `********` |
| `/api/projects/:id/env/:key` | DELETE | ✅ Delete single env var |
| `/api/projects/:id/deploy` | POST | ✅ Queue deploy job |
| `/api/projects/:id/deployments` | GET | ✅ List deployment history |
| `/api/deployments/:id` | GET | ✅ Deployment detail with commit info |
| `/api/deployments/:id/logs` | GET | ✅ SSE log stream (clone→detect→build steps) |
| `/api/detect` | POST | ✅ Auto-detect language/framework |

**Deploy Pipeline:** clone → detect → build → deploy → healthcheck (all steps functional)
**Deployed:** Docker image `deploy-mypaas:latest` (89.4MB) running on Ubuntu VM

**Deliverable:** ✅ Có thể tạo project, build, deploy, xem logs qua API.

### Phase 2: Frontend UI ✅ DONE & TESTED (2026-04-08)

```
✅ React 19 + Vite 6 + Tailwind CSS 3.4
✅ Dashboard (list projects, status, health indicator)
✅ Project detail (deployments, logs, env, settings tabs)
✅ Deploy button + realtime log viewer (SSE streaming)
✅ Env editor với "Ready to Deploy" flow (mask secrets, add/delete/save)
✅ Service management (create/start/stop/delete Postgres/Redis/MySQL/Mongo/MinIO)
✅ Service linking (auto inject DATABASE_URL, REDIS_URL etc.)
✅ Basic metrics (container CPU/memory/network per container)
✅ Services page with create dialog, status, stats display
✅ System stats summary (total CPU, memory, container count)
```

**Phase 2 API Test Results (2026-04-08):**
| API Endpoint | Method | Status |
|---|---|---|
| `/api/services` | GET | ✅ List all services |
| `/api/services` | POST | ✅ Create service (auto default image) |
| `/api/services/:id/start` | POST | ✅ Pull image + start container |
| `/api/services/:id/stop` | POST | ✅ Stop container |
| `/api/services/:id` | DELETE | ✅ Stop + delete service + links |
| `/api/services/:id/link/:projectId` | POST | ✅ Link + auto inject env vars (DATABASE_URL etc.) |
| `/api/services/:id/link/:projectId` | DELETE | ✅ Unlink service |
| `/api/stats` | GET | ✅ CPU%, memory, network per container |
| `/api/projects/:id/stats` | GET | ✅ Per-project container stats |

**Deliverable:** ✅ UI hoàn chỉnh, Services page, Stats dashboard, user có thể thao tác toàn bộ qua browser.

### Phase 3: CI/CD Automation ✅ DONE & TESTED (2026-04-08)

```
✅ GitHub webhook receiver (HMAC SHA-256 signature verification)
✅ Auto deploy on push (match by repo URL + branch + auto_deploy flag)
✅ Watcher system (poll git ls-remote every 60s, detect new commits)
✅ Env change → ready_to_deploy flow (already in Phase 1)
✅ Blue-green deployment (start new → healthcheck → stop old)
✅ Auto rollback on health check failure (in deploy worker)
✅ Deploy history + rollback UI (rollback button per deployment)
✅ Webhook URL display + copy button in project settings
```

**Phase 3 API Test Results (2026-04-08):**
| API Endpoint | Method | Status |
|---|---|---|
| `/api/webhooks/github` | POST (ping) | ✅ Returns `{"message":"pong"}` |
| `/api/webhooks/github` | POST (push) | ✅ Match project by repo URL+branch, trigger deploy |
| `/api/webhooks/github` | POST (unknown repo) | ✅ Returns `no matching project found` |
| `/api/deployments/:id/rollback` | POST | ✅ Returns 400 if no image (correct), 202 if rollback-able |

**Deliverable:** ✅ Full CI/CD pipeline — webhook auto-deploy, git polling watcher, rollback UI.

### Phase 4: Production Ready ✅ DONE & TESTED (2026-04-08)

```
✅ Traefik integration (reverse proxy + auto SSL) — Traefik v2.11, Docker labels, port 80/443
✅ Custom domain support — CRUD API, Traefik routers with optional Let's Encrypt SSL
✅ Service linking (auto inject DATABASE_URL) — moved from Phase 4, done in Phase 2
✅ Build caching (Docker layer cache) — CacheFrom + tag :latest for layer reuse
✅ Resource limits (CPU/RAM per container) — NanoCPUs + Memory in RunContainer
⏭️ Drag-drop canvas — skipped (complex UI, not production-essential)
✅ Authentication (bcrypt + session tokens) — setup/login/logout, 30-day sessions, auth middleware
```

**Phase 4 Implementation Details:**

| Feature | Files | Notes |
|---------|-------|-------|
| **Authentication** | `model/user.go`, `handler/auth.go`, `middleware/auth.go` | bcrypt hashing, crypto/rand tokens, Bearer + query param for SSE |
| **Traefik** | `deploy/docker-compose.yml`, `worker/deploy.go` | Traefik v2.11 (Docker provider), auto labels on project containers |
| **Custom Domains** | `model/domain.go`, `handler/domain.go`, `store/store.go` | CRUD, Traefik routers with optional SSL per domain |
| **Build Cache** | `docker/client.go`, `worker/deploy.go` | `CacheFrom` + `TagImage :latest` for Docker layer reuse |
| **Resource Limits** | `model/project.go`, `docker/client.go`, `worker/deploy.go` | `NanoCPUs` + `Memory` in container HostConfig |

**Phase 4 API Test Results (2026-04-08):**
| API Endpoint | Method | Status |
|---|---|---|
| `/api/auth/status` | GET | ✅ Returns `setup_required: true` when no users |
| `/api/auth/setup` | POST | ✅ Create admin user + session token (bcrypt) |
| `/api/auth/login` | POST | ✅ Returns token + user + expiry |
| `/api/auth/logout` | POST | ✅ Invalidates session token |
| `/api/projects` (no token) | GET | ✅ Returns 401 Unauthorized |
| `/api/projects` (with token) | GET | ✅ Returns projects list |
| `/api/projects/:id` | PUT | ✅ cpu_limit=0.5, mem_limit=256 saved correctly |
| `/api/projects/:id/domains` | POST | ✅ Add custom domain with ssl_auto |
| `/api/projects/:id/domains` | GET | ✅ List domains for project |
| `/api/domains/:id` | DELETE | ✅ Delete domain |
| Traefik routing (port 80) | GET | ✅ Routes to mypaas-server via Docker labels |

**Deliverable:** ✅ Production-ready self-hosted PaaS with auth, reverse proxy, custom domains, caching, and resource limits.

### Phase 5: Scale & Operations

> **Mục tiêu:** Nâng cấp từ single-node → production cluster, thêm quản lý dữ liệu và team.

**Phân tích hạ tầng hiện tại (2026-04-08):**
| Resource | Current | Constraint |
|----------|---------|------------|
| VM | Ubuntu 24.04, VirtualBox NAT (enp0s3: 10.0.2.15) | Single node |
| CPU | 2 cores | Shared with Docker builds |
| RAM | 1.9GB (559MB used, 1.4GB available) | Tight for Swarm overhead |
| Disk | 12GB (1.2GB free, 90% used!) | **CRITICAL** — needs expansion |
| Docker | 29.4.0, Swarm state: inactive | Ready for `docker swarm init` |
| Network | NAT only, enp0s8 (Host-Only) present but DOWN | Need Host-Only or Internal for Swarm |
| Containers | mypaas-server + mypaas-traefik | Bridge network `mypaas-network` |

---

#### Phase 5a: VM Infrastructure (Chuẩn bị) ✅ DONE (2026-04-09)

```
✅ Expand disk VM hiện tại (12GB → 30GB)
  - VBoxManage modifyhd → lvextend → resize2fs
✅ Enable enp0s8 (Host-Only Adapter) trên VM hiện tại
  - Static IP: 192.168.56.10
  - Dùng cho inter-node communication (Swarm)
✅ Clone VM → Worker node
  - Static IP: 192.168.56.11
  - enp0s3 (NAT) giữ nguyên cho internet
  - enp0s8 (Host-Only) cho Swarm cluster
✅ Port forwarding mới trên Host:
  - VM1 (Manager): ssh:2222, api:8080, http:80, https:443
  - VM2 (Worker): ssh:2223 (chỉ cần SSH để quản lý)
✅ Verify connectivity: VM1 ↔ VM2 qua 192.168.56.x
```

#### Phase 5b: Docker Swarm Cluster ✅ DONE & TESTED (2026-04-09)

```
✅ Init Swarm trên VM1 (Manager):    docker swarm init --advertise-addr 192.168.56.10
✅ Join Swarm trên VM2 (Worker):     docker swarm join --token <token> 192.168.56.10:2377
✅ Convert docker-compose.yml → docker stack deploy (Swarm mode)
✅ Overlay network thay bridge:      mypaas-network → overlay driver
✅ Migrate server/docker/client.go:
  - Detect Swarm mode (docker info → Swarm.LocalNodeState)
  - RunContainer → Docker Service API (replicas, placement)
  - Service update (rolling update strategy)
  - Placement constraints (manager vs worker nodes)
  - ListSwarmServices, ListSwarmServiceTasks, SwarmManagerAddr, UpdateSwarmServiceLabels
✅ Migrate worker/deploy.go:
  - Deploy as Swarm service thay vì standalone container
  - Labels cho Traefik vẫn giữ nguyên (Docker provider)
  - Health check → service converge check
  - Rollback supports Swarm mode (UpdateSwarmService)
✅ Traefik config update:
  - --providers.docker.swarmMode=true
  - Labels prefix: traefik.http.services → deploy labels on service
✅ Frontend: Show node info, service replicas, placement
  - Swarm.tsx: services list, task placement per node, refresh, manager addr
```

**Schema changes cho Swarm:**
```sql
ALTER TABLE projects ADD COLUMN replicas INTEGER DEFAULT 1;
ALTER TABLE projects ADD COLUMN placement TEXT DEFAULT ''; -- node constraint
```

**Kiến trúc Swarm:**
```
┌─────────────────────────────┐    ┌─────────────────────────────┐
│   VM1 - Manager (Leader)    │    │   VM2 - Worker              │
│   192.168.56.10             │    │   192.168.56.11             │
│                             │    │                             │
│   ┌─────────────────────┐   │    │   ┌─────────────────────┐   │
│   │ mypaas-server       │   │    │   │ app replicas        │   │
│   │ mypaas-traefik      │   │    │   │ (scheduled by Swarm)│   │
│   │ SQLite DB           │   │    │   │                     │   │
│   └─────────────────────┘   │    │   └─────────────────────┘   │
│                             │    │                             │
│   Swarm Manager + Raft      │◄──►│   Swarm Worker              │
│   Port 2377/7946/4789       │    │   Port 7946/4789            │
└─────────────────────────────┘    └─────────────────────────────┘
          ▲ Host-Only Network (enp0s8) ▲
```

#### Phase 5c: Persistent Volume Management ✅ DONE & TESTED (2026-04-09)

```
✅ API cho volume CRUD:
  - POST /api/projects/:id/volumes   — Create named volume
  - GET  /api/projects/:id/volumes   — List volumes
  - DELETE /api/projects/:id/volumes/:volumeId — Delete volume
✅ Model: volumes table (id, name, mount_path, project_id)
✅ worker/deploy.go: Pass volume mounts vào RunContainer/ServiceCreate
  - Container mode: Binds (mypaas-{name}-{vol}:{path})
  - Swarm mode: mount.Mount with TypeVolume
✅ UI: Volume management trong Project Detail (Volumes tab)
  - Create/delete volumes with name + mount path
  - Volume count badge on tab
✅ Audit logging cho volume create/delete
```

#### Phase 5d: Backup & Restore ✅ DONE & TESTED (2026-04-09)

```
✅ SQLite backup:
  - POST /api/backups              — sqlite3 .backup → /data/backups/
  - GET  /api/backups              — List backup files
  - GET  /api/backups/:id/download — Download backup file
  - POST /api/backups/:id/restore  — Restore from backup
  - DELETE /api/backups/:id        — Delete backup
✅ Admin-only access for create/restore/delete (RBAC enforced)
✅ UI: Backups page with create, download, restore, delete
✅ Audit logging cho backup operations
✅ Export/Import project: covered via backup/restore
```

#### Phase 5e: Team Collaboration & RBAC ✅ DONE & TESTED (2026-04-09)

```
✅ User roles: admin | member | viewer
✅ RoleRequired middleware: enforces role-based access on admin routes
✅ User management: list, update role, delete (admin only)
✅ Invite system:
  - POST /api/invitations          — Admin invites user (email + role, 7-day expiry)
  - GET  /api/invitations          — List all invitations
  - POST /api/auth/register        — Accept invite (public, token-based)
✅ Registration page: /register/:token with username/password form
  - Auto-login after registration
  - Copy invite link button in Users page
✅ Audit log:
  - Table: audit_logs (user_id, username, action, resource, resource_id, details)
  - Tracks: deploy, env change, project CRUD, user actions, backup, volume ops
✅ UI: Users page with 3 tabs (Users, Invitations, Audit Log)
✅ RBAC enforced: member/viewer blocked from users, backups create/restore, swarm init
```

#### Phase 5f: Marketplace (One-Click Templates) ✅ DONE & TESTED (2026-04-09)

```
✅ Template registry: 6 embedded templates
  - WordPress (PHP + MySQL)
  - Node.js + PostgreSQL
  - Redis Cache
  - Node.js + Redis
  - PostgreSQL standalone
  - MinIO Object Storage
✅ GET  /api/marketplace           — List all templates
✅ POST /api/marketplace/:id/deploy — Deploy template (creates projects + services)
✅ UI: Marketplace page with grid layout, deploy button, name prompt
✅ Audit logging cho template deployments
```

**Thứ tự triển khai đề xuất:**
1. ✅ **5a** — VM infra (disk + network) — DONE
2. ✅ **5b** — Docker Swarm — DONE
3. ✅ **5c** — Persistent Volumes — DONE
4. ✅ **5d** — Backup/Restore — DONE
5. ✅ **5e** — Team/RBAC — DONE
6. ✅ **5f** — Marketplace — DONE

---

## 10. Tech Stack tổng hợp

| Component | Technology | Lý do |
|-----------|-----------|-------|
| **API Server** | Go + Fiber | Tận dụng `core/` module, performance, single binary |
| **Database** | SQLite (→ Postgres later) | Zero-config, embedded, đủ cho single-node |
| **Queue** | Go channels + goroutines | Đơn giản, không cần Redis queue ban đầu |
| **Frontend** | React + Vite + Tailwind + shadcn/ui | Modern, fast, reuse từ extension webview |
| **Docker** | Docker Engine API (Go SDK) | Native Go SDK, không qua CLI |
| **Proxy** | Traefik v2.11 | Auto-discovery, Docker labels, auto SSL |
| **Logs** | SSE (Server-Sent Events) | Đơn giản hơn WebSocket, đủ cho log streaming |
| **Build** | Docker BuildKit | Tận dụng từ railpack patterns |

---

## 11. Cấu trúc thư mục dự án mới

```
my-paas/
├── core/                    # ← GIỮ NGUYÊN - auto-detect + dockerfile gen
├── server/                  # ← MỚI - API server + workers
│   ├── main.go
│   ├── go.mod
│   ├── handler/
│   ├── model/
│   ├── worker/
│   ├── docker/
│   ├── proxy/
│   ├── store/
│   └── watcher/
├── web/                     # ← MỚI - Frontend UI
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── App.tsx
│   │   ├── pages/
│   │   ├── components/
│   │   └── api/
│   └── public/
├── deploy/                  # ← MỚI - Docker configs để deploy My PaaS
│   ├── Dockerfile           # Multi-stage: build Go + React → single image
│   ├── docker-compose.yml   # My PaaS + Traefik
│   └── traefik/
├── my-paas-extension/       # ← GIỮ NGUYÊN - VS Code extension (optional)
├── railpack/                # ← GIỮ NGUYÊN - reference
└── docs/
```

---

## 12. Dockerfile cho My PaaS service

```dockerfile
# Stage 1: Build Frontend
FROM node:22-alpine AS frontend
WORKDIR /build
COPY web/package.json web/bun.lockb ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build API Server
FROM golang:1.26-alpine AS backend
WORKDIR /build
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY core/ ../core/
RUN CGO_ENABLED=1 go build -ldflags="-w -s" -o /mypaas .

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates git docker-cli
WORKDIR /app
COPY --from=backend /mypaas .
COPY --from=frontend /build/dist ./static/
EXPOSE 8080
CMD ["./mypaas"]
```

---

## 13. Quick Start (user perspective)

```bash
# 1. Cài đặt My PaaS trên VPS
curl -sSL https://mypaas.dev/install.sh | sh
# Hoặc:
docker compose -f docker-compose.mypaas.yml up -d

# 2. Truy cập UI
open http://your-server:3000

# 3. Tạo project mới
#    - Nhập Git URL: https://github.com/user/my-app.git
#    - Auto detect: "Node.js / Next.js"
#    - Thêm env vars: DATABASE_URL, API_KEY...
#    - Click "Deploy"

# 4. Sau khi deploy xong:
#    - App chạy tại: https://my-app.your-domain.com
#    - Logs realtime trong UI
#    - Push code → auto deploy
#    - Sửa env → badge "Ready to Deploy" → click Deploy
```

---

## Tóm tắt

Dự án này là bước tiến từ:
- **v1 (core/)**: Library detect + generate Dockerfile
- **v2 (extension/)**: VS Code drag-drop designer
- **v3 (server/)**: Full self-hosted PaaS giống Railway

Ưu điểm so với Railway:
- **Self-hosted**: Data hoàn toàn của bạn
- **Single Docker service**: Dễ cài, dễ chạy
- **Tận dụng core/**: Auto-detect đã có sẵn
- **Không cần Kubernetes**: Docker Engine là đủ cho hầu hết use cases
