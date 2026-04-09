# Quản lý dự án

## Tổng quan

Project là đơn vị cốt lõi trong My PaaS — mỗi project tương ứng với một ứng dụng web được triển khai từ Git repository.

## Vòng đời project

```
Create → Configure → Deploy → Monitor → Scale → Destroy
```

### Tạo project

```bash
# CLI
mypaas create --name my-app --git-url https://github.com/user/repo.git --branch main

# API
POST /api/projects
{
  "name": "my-app",
  "git_url": "https://github.com/user/repo.git",
  "branch": "main"
}
```

### Auto-detection

Khi deploy, hệ thống tự nhận diện ngôn ngữ và framework:

| Ngôn ngữ | File đặc trưng | Framework | Base image |
|---|---|---|---|
| **Node.js** | `package.json` | Next.js, React, Express, Vite | `node:22-alpine` |
| **Python** | `requirements.txt`, `Pipfile` | Django, Flask, FastAPI | `python:3.12-slim` |
| **Go** | `go.mod` | Gin, Fiber, Echo | `golang:1-alpine` |
| **Java** | `pom.xml`, `build.gradle` | Spring Boot | `eclipse-temurin:21` |
| **PHP** | `composer.json` | Laravel, Symfony | `php:8.2-apache` |
| **Rust** | `Cargo.toml` | Actix, Rocket | `rust:1-alpine` |
| **Static** | `index.html` | — | `nginx:alpine` |

### Deployment triggers

| Trigger | Cách sử dụng |
|---|---|
| **Manual** | Click "Deploy" trên Dashboard hoặc `mypaas deploy <id>` |
| **Git Webhook** | Push lên GitHub → `POST /api/webhooks/github` |
| **Auto-deploy** | Bật `auto_deploy` → Git Watcher poll mỗi 60s |
| **Env change** | Thay đổi biến môi trường → auto-redeploy |

### Environment variables

```bash
# CLI
mypaas env <project-id> set DATABASE_URL=postgres://... NODE_ENV=production
mypaas env <project-id>               # Liệt kê
mypaas env <project-id> delete KEY    # Xoá

# API
PUT /api/projects/:id/env
{
  "vars": [
    {"key": "DATABASE_URL", "value": "postgres://...", "is_secret": true},
    {"key": "NODE_ENV", "value": "production"}
  ]
}
```

Secret vars (`is_secret: true`) được mã hoá AES-256-GCM trước khi lưu.

### Custom domains

```bash
# API
POST /api/projects/:id/domains
{ "domain": "app.company.com", "ssl_auto": true }
```

Traefik tự động cấu hình routing và Let's Encrypt SSL.

### Resource limits

```json
{
  "cpu_limit": 1.0,
  "mem_limit": 536870912,
  "replicas": 2
}
```

### Container actions

| Action | API | Mô tả |
|---|---|---|
| **Start** | `POST /projects/:id/start` | Khởi động container/service |
| **Stop** | `POST /projects/:id/stop` | Dừng (scale to 0 nếu Swarm) |
| **Restart** | `POST /projects/:id/restart` | Force restart |
| **Rollback** | `POST /deployments/:id/rollback` | Quay lại image của deployment trước |

### Monitoring

- **Real-time logs**: `GET /projects/:id/logs` (SSE stream)
- **Container stats**: `GET /projects/:id/stats` (CPU%, Memory, Network I/O)
- **Deployment history**: `GET /projects/:id/deployments`

## Persistent Volumes

```bash
POST /api/projects/:id/volumes
{ "name": "app-data", "mount_path": "/data" }
```

Docker volume được mount vào container, dữ liệu persist qua restarts và redeploys.
