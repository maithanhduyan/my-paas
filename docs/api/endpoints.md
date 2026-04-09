# API Reference

## Base URL

```
https://your-server.com/api
```

## Authentication

Hầu hết endpoint yêu cầu authentication qua header:

```
Authorization: Bearer <jwt-token>
```

hoặc API Key:

```
X-API-Key: mpk_live_xxxxx
```

## Endpoints

### Health & System

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/health` | — | Health check |
| GET | `/metrics` | — | Prometheus metrics |
| GET | `/docs/openapi.json` | — | OpenAPI spec (JSON) |
| GET | `/docs` | — | Swagger UI |

### Auth

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/auth/status` | — | Kiểm tra trạng thái setup |
| POST | `/auth/setup` | — | Initial admin setup |
| POST | `/auth/login` | — | Đăng nhập |
| POST | `/auth/register` | — | Đăng ký bằng invitation token |
| POST | `/auth/logout` | ✅ | Đăng xuất |
| POST | `/auth/refresh` | ✅ | Refresh JWT token |

### Projects

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/projects` | ✅ | Liệt kê tất cả projects |
| POST | `/projects` | ✅ | Tạo project mới |
| GET | `/projects/:id` | ✅ | Chi tiết project |
| PUT | `/projects/:id` | ✅ | Cập nhật project |
| DELETE | `/projects/:id` | ✅ | Xoá project |
| POST | `/detect` | ✅ | Detect ngôn ngữ project |

### Deployments

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| POST | `/projects/:id/deploy` | ✅ | Trigger deployment |
| GET | `/projects/:id/deployments` | ✅ | Lịch sử deployments |
| GET | `/deployments/:id` | ✅ | Chi tiết deployment |
| POST | `/deployments/:id/rollback` | ✅ | Rollback deployment |

### Container Actions

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| POST | `/projects/:id/start` | ✅ | Khởi động container |
| POST | `/projects/:id/stop` | ✅ | Dừng container |
| POST | `/projects/:id/restart` | ✅ | Restart container |

### Environment

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/projects/:id/env` | ✅ | Liệt kê env vars |
| PUT | `/projects/:id/env` | ✅ | Cập nhật env vars |
| DELETE | `/projects/:id/env/:key` | ✅ | Xoá env var |

### Logs (SSE)

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/projects/:id/logs` | ✅ | Stream logs (Server-Sent Events) |
| GET | `/deployments/:id/logs` | ✅ | Stream deployment logs |

### Services

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/services` | ✅ | Liệt kê services |
| POST | `/services` | ✅ | Tạo service |
| DELETE | `/services/:id` | ✅ | Xoá service |
| POST | `/services/:id/start` | ✅ | Khởi động service |
| POST | `/services/:id/stop` | ✅ | Dừng service |
| POST | `/services/:id/link/:projectId` | ✅ | Link service ↔ project |
| DELETE | `/services/:id/link/:projectId` | ✅ | Unlink service |

### Domains

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/projects/:id/domains` | ✅ | Liệt kê domains |
| POST | `/projects/:id/domains` | ✅ | Thêm domain |
| DELETE | `/domains/:id` | ✅ | Xoá domain |

### Stats

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/stats` | ✅ | System-wide stats |
| GET | `/projects/:id/stats` | ✅ | Container stats (CPU, Memory) |

### Volumes

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/projects/:id/volumes` | ✅ | Liệt kê volumes |
| POST | `/projects/:id/volumes` | ✅ | Tạo volume |
| DELETE | `/projects/:id/volumes/:volumeId` | ✅ | Xoá volume |

### Backups

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/backups` | ✅ | Liệt kê backups |
| POST | `/backups` | 🔒 Admin | Tạo backup |
| GET | `/backups/:id/download` | ✅ | Download backup |
| POST | `/backups/:id/restore` | 🔒 Admin | Restore backup |
| DELETE | `/backups/:id` | 🔒 Admin | Xoá backup |

### Marketplace

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/marketplace` | ✅ | Liệt kê templates |
| POST | `/marketplace/:id/deploy` | ✅ | Deploy template |

### Samples

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/samples` | ✅ | Liệt kê sample apps |

### Users

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/users` | 🔒 Admin | Liệt kê users |
| PUT | `/users/:id/role` | 🔒 Admin | Thay đổi role |
| DELETE | `/users/:id` | 🔒 Admin | Xoá user |
| POST | `/invitations` | 🔒 Admin | Mời user mới |
| GET | `/invitations` | 🔒 Admin | Liệt kê invitations |
| GET | `/audit` | 🔒 Admin | Xem audit logs |

### Swarm

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/swarm/status` | ✅ | Trạng thái Swarm |
| GET | `/swarm/services` | ✅ | Liệt kê Swarm services |
| POST | `/swarm/init` | 🔒 Admin | Khởi tạo Swarm |
| GET | `/swarm/token` | 🔒 Admin | Lấy join token |

### Registry

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/registry/status` | ✅ | Trạng thái registry |
| POST | `/registry/start` | 🔒 Admin | Khởi động registry |
| POST | `/registry/stop` | 🔒 Admin | Dừng registry |
| GET | `/registry/images` | ✅ | Liệt kê images |
| DELETE | `/registry/images/:name` | 🔒 Admin | Xoá image |
| POST | `/projects/:id/push` | ✅ | Push image lên registry |

### Organizations

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/organizations` | ✅ | Liệt kê orgs |
| POST | `/organizations` | ✅ | Tạo org |
| GET | `/organizations/:id` | ✅ | Chi tiết org |
| PUT | `/organizations/:id` | ✅ | Cập nhật org |
| DELETE | `/organizations/:id` | 🔒 Admin | Xoá org |
| GET | `/organizations/:id/members` | ✅ | Liệt kê members |
| POST | `/organizations/:id/members` | ✅ | Thêm member |
| DELETE | `/organizations/:id/members/:userId` | ✅ | Xoá member |
| GET | `/organizations/:id/quotas` | ✅ | Xem quota usage |

### API Keys

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/api-keys` | ✅ | Liệt kê API keys |
| POST | `/api-keys` | ✅ | Tạo API key |
| DELETE | `/api-keys/:id` | ✅ | Xoá API key |

### Webhooks

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| POST | `/webhooks/github` | — | GitHub webhook receiver |

### Notifications

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/notifications/channels` | ✅ | Liệt kê channels |
| POST | `/notifications/channels` | ✅ | Tạo channel |
| PUT | `/notifications/channels/:id` | ✅ | Cập nhật channel |
| DELETE | `/notifications/channels/:id` | ✅ | Xoá channel |
| GET | `/notifications/channels/:channelId/rules` | ✅ | Liệt kê rules |
| POST | `/notifications/channels/:channelId/rules` | ✅ | Tạo rule |
| DELETE | `/notifications/rules/:ruleId` | ✅ | Xoá rule |

---

**Tổng cộng: 76 endpoints** — 8 public, 14 admin-only, 54 authenticated.

::: tip Swagger UI
Truy cập `/api/docs` để xem API documentation tương tác với Swagger UI.
:::
