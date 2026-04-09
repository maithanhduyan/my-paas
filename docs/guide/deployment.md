# Cài đặt & Triển khai

## Yêu cầu hệ thống

| Thành phần | Yêu cầu tối thiểu | Khuyến nghị |
|---|---|---|
| **OS** | Linux (Ubuntu 22.04+) | Ubuntu 24.04 LTS |
| **Docker** | 24.0+ | 27.0+ |
| **RAM** | 1 GB | 4 GB |
| **Disk** | 10 GB | 40 GB SSD |
| **CPU** | 1 vCPU | 2 vCPU |

## Cài đặt nhanh (Standalone)

```bash
# 1. Clone repository
git clone https://github.com/my-paas/my-paas.git
cd my-paas

# 2. Khởi chạy
docker compose -f deploy/docker-compose.yml up -d

# 3. Truy cập
# http://<server-ip>:8080
```

Lần đầu truy cập → tạo tài khoản admin → bắt đầu deploy ứng dụng.

## Docker Swarm (Multi-node)

```bash
# Node 1 (Manager)
docker swarm init --advertise-addr <MANAGER_IP>

# Node 2+ (Worker) — chạy lệnh join từ output ở trên
docker swarm join --token <TOKEN> <MANAGER_IP>:2377

# Deploy stack
docker stack deploy -c deploy/docker-stack.yml mypaas
```

## Enterprise Stack

Bao gồm PostgreSQL, Redis, Prometheus:

```bash
docker stack deploy -c deploy/docker-stack.enterprise.yml mypaas
```

Biến môi trường enterprise cần thiết:

```env
MYPAAS_DB_DRIVER=postgres
MYPAAS_DB_URL=postgres://mypaas:secret@postgres:5432/mypaas?sslmode=disable
MYPAAS_REDIS_URL=redis://redis:6379
MYPAAS_SECRET=<32+ byte secret key>
MYPAAS_ENCRYPTION_KEY=<64 hex chars for AES-256>
```

## Build từ source

```bash
# Backend
cd server
go build -ldflags="-w -s" -o mypaas-server .

# Frontend
cd web
npm ci && npm run build
# Output: web/dist/ → copy vào server/static/

# CLI
cd cli
go build -o mypaas .
```

## Docker image tự build

```bash
docker compose -f deploy/docker-compose.yml build
```

Dockerfile sử dụng 3-stage build:

1. **`node:22-alpine`** — build React frontend
2. **`golang:1-alpine`** — compile Go server (CGO enabled cho SQLite)
3. **`alpine:3.20`** — runtime image ~30MB

## Health check

```bash
curl http://localhost:8080/api/health
# {"status":"ok","docker":"connected","go":"go1.26.2"}
```
