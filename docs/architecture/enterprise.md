# Enterprise Features

My PaaS v4.0.0 bổ sung tầng enterprise-grade lên trên core platform, cho phép sử dụng trong môi trường production với yêu cầu bảo mật, giám sát và quản lý đội nhóm.

## Authentication Triple-Layer

Hệ thống hỗ trợ 3 phương thức xác thực, kiểm tra theo thứ tự:

```
Request → JWT? → API Key? → Session Token? → 401 Unauthorized
```

### JWT (JSON Web Token)

- **Algorithm**: HMAC-SHA256 (symmetric)
- **Access token**: TTL 24 giờ, chứa `sub`, `username`, `role`, `type`
- **Refresh token**: TTL 30 ngày, dùng để lấy access token mới
- **Custom implementation**: không dependency ngoài, dùng `crypto/hmac` + `encoding/base64`

```
POST /api/auth/login
  → { access_token, refresh_token, expires_in, user }

POST /api/auth/refresh
  → { access_token, expires_in }
```

### API Key

- **Format**: `mpk_live_<random>` (prefix cho dễ nhận diện)
- **Storage**: SHA-256 hash — key gốc chỉ hiển thị 1 lần khi tạo
- **Scopes**: comma-separated (`read`, `deploy`, `admin`, `*`)
- **Use case**: CI/CD pipeline, automation scripts

### Session Token

- **Backward compatible** với auth cũ (pre-enterprise)
- **Database-backed**: table `sessions` với `token` unique, `expires_at`
- **Dùng cho**: web dashboard khi JWT chưa enable

## Organizations (Multi-tenant)

### Cấu trúc

```
Organization
├── Members (owner, admin, member, viewer)
├── Projects (thuộc org)
├── Resource Quotas
│   ├── max_projects
│   ├── max_services
│   ├── max_cpu (cores)
│   ├── max_memory (bytes)
│   └── max_deployments (per month)
└── Notification Channels
```

### Phân quyền

| Role | Quyền |
|---|---|
| `owner` | Toàn quyền, xoá org, quản lý billing |
| `admin` | Quản lý members, projects, settings |
| `member` | Tạo/deploy projects, manage services |
| `viewer` | Xem logs, stats — không thay đổi |

### Quota Enforcement

API `GET /organizations/:id/quotas` trả về usage hiện tại so với limits:

```json
{
  "max_projects": 50,
  "used_projects": 12,
  "max_services": 20,
  "used_services": 3,
  "max_deployments": 500,
  "used_deployments": 87
}
```

## Encryption

### AES-256-GCM (Secret Encryption)

Env vars đánh dấu `is_secret = true` được mã hoá trước khi lưu:

```
Plaintext → AES-256-GCM Encrypt(key, nonce, plaintext) → Base64 → DB
DB → Base64 decode → AES-256-GCM Decrypt(key, nonce, ciphertext) → Container ENV
```

- **Key**: 256-bit từ biến môi trường `MYPAAS_ENCRYPTION_KEY` (64 hex chars)
- **Nonce**: 12 bytes random, prepend vào ciphertext
- **Authenticated**: GCM mode đảm bảo integrity + confidentiality

### Password Hashing

```
Password → bcrypt(password, cost=10) → hash → DB
Login → bcrypt.CompareHashAndPassword(hash, input)
```

## Rate Limiting

Sliding window rate limiter với 2 backend:

### Redis-backed (distributed)

```
Key: ratelimit:<ip>:<window>
Mechanism: INCR + EXPIRE
```

Phù hợp khi chạy multiple replicas — mọi instance share Redis state.

### In-memory (fallback)

```
sync.Map[ip] → { count, windowStart }
Cleanup: goroutine mỗi 60s xoá expired entries
```

Tự động sử dụng khi Redis không available.

### Config

| Biến | Default | Mô tả |
|---|---|---|
| `MYPAAS_RATE_LIMIT_RPS` | 100 | Request/giây/IP |
| `MYPAAS_RATE_LIMIT_BURST` | 200 | Burst cap |

Response khi bị limit: `429 Too Many Requests`.

## Security Headers

Middleware `SecurityHeaders()` gắn vào mọi response:

| Header | Value | Mục đích |
|---|---|---|
| `X-Content-Type-Options` | `nosniff` | Chống MIME-type sniffing |
| `X-Frame-Options` | `DENY` | Chống clickjacking |
| `X-XSS-Protection` | `1; mode=block` | XSS filter (legacy) |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Enforce HTTPS |
| `Content-Security-Policy` | `default-src 'self'` | Chống XSS, injection |
| `X-Request-Id` | UUID | Request tracing |

## Prometheus Metrics

Endpoint: `GET /api/metrics` (public, text format)

### Metrics cung cấp

| Metric | Type | Mô tả |
|---|---|---|
| `mypaas_deployments_total` | Counter | Tổng số deployments |
| `mypaas_deployments_success` | Counter | Deployments thành công |
| `mypaas_deployments_failed` | Counter | Deployments thất bại |
| `mypaas_http_requests_total` | Counter | Tổng HTTP requests |
| `mypaas_http_request_duration_seconds` | Histogram | Latency phân bổ |
| `mypaas_active_projects` | Gauge | Số project đang chạy |
| `mypaas_queue_depth` | Gauge | Jobs đang chờ trong queue |
| `mypaas_uptime_seconds` | Gauge | Server uptime |
| `go_goroutines` | Gauge | Số goroutine active |
| `go_memstats_alloc_bytes` | Gauge | Memory đang sử dụng |

### Prometheus config

```yaml
# prometheus.yml
scrape_configs:
  - job_name: mypaas
    scrape_interval: 15s
    static_configs:
      - targets: ['mypaas:8080']
    metrics_path: /api/metrics
```

## Notification System

### Channels

| Type | Config | Delivery |
|---|---|---|
| `webhook` | `{"url": "https://..."}` | HTTP POST JSON |
| `slack` | `{"webhook_url": "https://hooks.slack.com/..."}` | Slack Incoming Webhook |
| `email` | `{"smtp_host": "...", "to": "..."}` | SMTP (reserved) |

### Event Rules

Mỗi channel có rules xác định events nào được gửi:

| Event | Trigger |
|---|---|
| `deploy.started` | Deployment bắt đầu |
| `deploy.succeeded` | Deployment thành công |
| `deploy.failed` | Deployment thất bại |
| `health.down` | Container/service không healthy |
| `health.recovered` | Container/service recovered |
| `quota.warning` | Sắp hết quota (80%+) |
| `backup.completed` | Backup hoàn tất |

## Audit Logging

Mọi hành động admin được ghi nhận:

```json
{
  "id": "...",
  "user_id": "0e572f16",
  "username": "admin",
  "action": "create_project",
  "resource": "project",
  "resource_id": "abc123",
  "details": "{\"name\":\"my-app\"}",
  "created_at": "2026-04-09T..."
}
```

Truy vấn: `GET /api/audit` (admin only).
