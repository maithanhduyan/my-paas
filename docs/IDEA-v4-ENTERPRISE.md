# My PaaS v4 - Enterprise Class

> Nâng cấp My PaaS thành nền tảng PaaS cấp doanh nghiệp — cho phép tổ chức tự triển khai cloud PaaS riêng trên server của họ. Hoạt động độc lập, ổn định, hiệu suất cao.

---

## 1. Tổng quan Enterprise Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     ENTERPRISE MY-PAAS CLUSTER                         │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                        LOAD BALANCER                              │  │
│  │                    (Traefik HA / HAProxy)                         │  │
│  └───────────────────┬───────────────────┬───────────────────────────┘  │
│                      │                   │                              │
│  ┌───────────────────┴───┐   ┌──────────┴────────────┐                 │
│  │   API Server (Go)     │   │   API Server (Go)     │   ← Horizontal │
│  │   + JWT Auth          │   │   + JWT Auth          │     Scale       │
│  │   + Rate Limiting     │   │   + Rate Limiting     │                 │
│  │   + Prometheus        │   │   + Prometheus        │                 │
│  └───────────┬───────────┘   └───────────┬───────────┘                 │
│              │                           │                              │
│  ┌───────────┴───────────────────────────┴───────────────────────────┐  │
│  │                    SHARED INFRASTRUCTURE                          │  │
│  │                                                                   │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────────┐ │  │
│  │  │PostgreSQL│ │  Redis   │ │ Docker   │ │ S3/MinIO             │ │  │
│  │  │  (HA)    │ │ Cluster  │ │ Registry │ │ (Artifacts/Backups)  │ │  │
│  │  │          │ │ (Queue,  │ │ (Local)  │ │                      │ │  │
│  │  │          │ │  Cache,  │ │          │ │                      │ │  │
│  │  │          │ │  PubSub) │ │          │ │                      │ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                    DOCKER SWARM CLUSTER                           │  │
│  │                                                                   │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │  │
│  │  │Manager 1│ │Manager 2│ │Manager 3│ │Worker 1 │ │Worker N │   │  │
│  │  │  (Raft) │ │  (Raft) │ │  (Raft) │ │         │ │         │   │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │  │
│  │                                                                   │  │
│  │  Running: App containers, Service containers, Build workers       │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                    MONITORING STACK                                │  │
│  │  Prometheus → Grafana → AlertManager → Webhook/Email/Slack        │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Enterprise Features Overview

| Category | Feature | Priority | Status |
|----------|---------|----------|--------|
| **Database** | PostgreSQL support (alongside SQLite) | P0 | ✅ |
| **Cache/Queue** | Redis for job queue, cache, pub/sub | P0 | ✅ |
| **Auth** | JWT tokens + API keys + refresh tokens | P0 | ✅ |
| **Security** | Rate limiting per IP/user/API key | P0 | ✅ |
| **Security** | Secrets encryption at rest (AES-256-GCM) | P0 | ✅ |
| **Security** | Security headers (HSTS, CSP, X-Frame) | P0 | ✅ |
| **Observability** | Prometheus metrics endpoint | P0 | ✅ |
| **Multi-tenancy** | Organizations + resource quotas | P1 | ✅ |
| **Notifications** | Webhook + Slack + Email alerts | P1 | ✅ |
| **Config** | Centralized config (env + file) | P0 | ✅ |
| **Operations** | Scheduled backups (cron) | P1 | ✅ |
| **Operations** | Health check dashboard | P1 | ✅ |
| **DX** | OpenAPI/Swagger documentation | P2 | — |
| **DX** | CLI tool (mypaas-cli) | P2 | — |
| **Registry** | Local Docker registry | P2 | — |

---

## 3. Implementation Details

### 3.1 Centralized Configuration (`server/config/config.go`)

```go
type Config struct {
    // Server
    Listen       string // :8080
    Domain       string // mypaas.company.com
    Secret       string // JWT signing + encryption key (32+ bytes)

    // Database (PostgreSQL or SQLite)
    DBDriver     string // "sqlite" | "postgres"
    DBPath       string // SQLite path (when driver=sqlite)
    DBURL        string // PostgreSQL connection URL

    // Redis
    RedisURL     string // redis://host:6379/0
    RedisEnabled bool

    // Security
    RateLimitRPS     int    // Requests per second per IP
    RateLimitBurst   int    // Burst size
    JWTExpiry        time.Duration
    RefreshExpiry    time.Duration
    EncryptionKey    string // 32-byte key for AES-256-GCM

    // Workers
    WorkerCount      int   // Concurrent deploy workers
    QueueSize        int   // Job queue buffer

    // Features
    RegistrationOpen bool  // Allow public registration
    MaintenanceMode  bool
}
```

Environment variables:
```bash
MYPAAS_LISTEN=:8080
MYPAAS_DOMAIN=mypaas.company.com
MYPAAS_SECRET=your-32-byte-secret-key-here!!!
MYPAAS_DB_DRIVER=postgres          # or "sqlite"
MYPAAS_DB_PATH=/data/mypaas.db    # SQLite only
MYPAAS_DB_URL=postgres://user:pass@host:5432/mypaas?sslmode=require
MYPAAS_REDIS_URL=redis://host:6379/0
MYPAAS_RATE_LIMIT_RPS=100
MYPAAS_RATE_LIMIT_BURST=200
MYPAAS_JWT_EXPIRY=24h
MYPAAS_REFRESH_EXPIRY=720h
MYPAAS_ENCRYPTION_KEY=32-byte-hex-key
MYPAAS_WORKER_COUNT=4
MYPAAS_QUEUE_SIZE=500
```

### 3.2 PostgreSQL Support

Database abstraction via `database/sql` driver swap — same SQL interface, different driver:

- **SQLite** (default, single-node): Zero config, embedded
- **PostgreSQL** (enterprise, multi-node): Connection pooling, concurrent writes, ACID

Migration differences handled with `$1` vs `?` placeholder adaptation.

### 3.3 Redis Integration

```
Redis roles:
├── Job Queue      — BList-based reliable queue (replaces Go channels)
├── Cache          — Session cache, project cache, stats cache
├── Pub/Sub        — Real-time log streaming across API instances
├── Rate Limiting  — Token bucket per IP/user (replaces in-memory)
└── Distributed Lock — Prevent concurrent deploys on same project
```

### 3.4 JWT Authentication

```
┌─────────┐         ┌─────────────┐         ┌──────────┐
│  Login   │────────>│  JWT Token  │────────>│  API     │
│          │         │  (15min)    │         │  Request │
└─────────┘         ├─────────────┤         └──────────┘
                    │  Refresh    │
                    │  Token      │
                    │  (30 days)  │
                    └─────────────┘

┌──────────────┐
│  API Key     │───> Long-lived, scoped tokens for CI/CD
│  (no expiry) │    Prefix: mpk_live_xxxxx / mpk_test_xxxxx
└──────────────┘
```

### 3.5 Organizations & Resource Quotas

```sql
CREATE TABLE organizations (
    id          TEXT PRIMARY KEY,
    name        TEXT UNIQUE NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    max_projects    INT DEFAULT 0,      -- 0 = unlimited
    max_services    INT DEFAULT 0,
    max_cpu         REAL DEFAULT 0,     -- Total CPU cores
    max_memory      BIGINT DEFAULT 0,   -- Total memory bytes
    max_deployments INT DEFAULT 0,      -- Per month
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE org_members (
    id      TEXT PRIMARY KEY,
    org_id  TEXT REFERENCES organizations(id),
    user_id TEXT REFERENCES users(id),
    role    TEXT DEFAULT 'member',  -- owner | admin | member | viewer
    UNIQUE(org_id, user_id)
);

-- Projects belong to org
ALTER TABLE projects ADD COLUMN org_id TEXT DEFAULT '' REFERENCES organizations(id);
```

### 3.6 Notification System

```go
type NotificationChannel string
const (
    ChannelWebhook NotificationChannel = "webhook"
    ChannelSlack   NotificationChannel = "slack"
    ChannelEmail   NotificationChannel = "email"
)

type NotificationEvent string
const (
    EventDeployStarted   NotificationEvent = "deploy.started"
    EventDeploySucceeded NotificationEvent = "deploy.succeeded"
    EventDeployFailed    NotificationEvent = "deploy.failed"
    EventHealthDown      NotificationEvent = "health.down"
    EventHealthRecovered NotificationEvent = "health.recovered"
    EventQuotaWarning    NotificationEvent = "quota.warning"
    EventBackupCompleted NotificationEvent = "backup.completed"
)
```

### 3.7 Prometheus Metrics

```
mypaas_deployments_total{status="healthy|failed",project="name"}
mypaas_deployments_duration_seconds{project="name"}
mypaas_active_projects_total
mypaas_active_services_total
mypaas_api_requests_total{method="GET|POST",path="/api/...",status="200|400|500"}
mypaas_api_request_duration_seconds{method,path}
mypaas_queue_depth
mypaas_worker_busy_count
mypaas_container_cpu_usage{project="name"}
mypaas_container_memory_usage{project="name"}
```

---

## 4. Deployment (Enterprise)

```yaml
# docker-compose.enterprise.yml
version: "3.8"
services:
  mypaas:
    image: mypaas:enterprise
    deploy:
      replicas: 2
      resources:
        limits: { cpus: "2", memory: "2G" }
    environment:
      MYPAAS_DB_DRIVER: postgres
      MYPAAS_DB_URL: postgres://mypaas:secret@postgres:5432/mypaas?sslmode=disable
      MYPAAS_REDIS_URL: redis://redis:6379/0
      MYPAAS_SECRET: ${MYPAAS_SECRET}
      MYPAAS_DOMAIN: ${MYPAAS_DOMAIN}
      MYPAAS_ENCRYPTION_KEY: ${MYPAAS_ENCRYPTION_KEY}
      MYPAAS_WORKER_COUNT: 4
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - mypaas-builds:/data/builds
    networks:
      - mypaas-network

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: mypaas
      POSTGRES_USER: mypaas
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - mypaas-network

  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    networks:
      - mypaas-network

  traefik:
    image: traefik:v2.11
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - traefik-certs:/certs
    command:
      - --providers.docker=true
      - --providers.docker.swarmMode=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}
      - --certificatesresolvers.letsencrypt.acme.storage=/certs/acme.json
      - --certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web
      - --metrics.prometheus=true
      - --metrics.prometheus.entrypoint=metrics
      - --entrypoints.metrics.address=:8082
    networks:
      - mypaas-network

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    networks:
      - mypaas-network

  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana-data:/var/lib/grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
    networks:
      - mypaas-network

volumes:
  postgres-data:
  redis-data:
  mypaas-builds:
  traefik-certs:
  prometheus-data:
  grafana-data:

networks:
  mypaas-network:
    driver: overlay
    attachable: true
```

---

## 5. Backward Compatibility

- **SQLite mode** remains default — zero-config for small teams
- **PostgreSQL mode** activated via `MYPAAS_DB_DRIVER=postgres`
- **Redis optional** — falls back to Go channels if `MYPAAS_REDIS_URL` empty
- **JWT auth** coexists with session tokens during migration
- **Organizations** optional — single-org mode when no org created

---

## 6. Migration Path

```
v3 (current)                    v4 (enterprise)
──────────────                  ─────────────────
SQLite only          →          SQLite + PostgreSQL
Go channel queue     →          Redis queue (+ fallback)
Session tokens       →          JWT + API keys (+ session compat)
No encryption        →          AES-256-GCM secrets
No rate limiting     →          Per-IP + per-user rate limiting
No metrics           →          Prometheus endpoint
Simple RBAC          →          Organizations + quotas
No notifications     →          Webhook/Slack/Email alerts
Single instance      →          Horizontally scalable
```
