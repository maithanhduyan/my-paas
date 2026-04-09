# Tiêu chuẩn bảo mật

## OWASP Top 10 Coverage

### A01 — Broken Access Control

- **JWT Authentication** — Token có expiry (24h access, 168h refresh)
- **API Key scoping** — Mỗi key có danh sách scopes giới hạn (`read`, `deploy`, `admin`, `*`)
- **Organization isolation** — Quota và resource limits per org
- **Admin-only endpoints** — Registry start/stop, user management, system backup chỉ admin truy cập được

### A02 — Cryptographic Failures

- **AES-256-GCM** — Encrypt secret environment variables tại rest
- **bcrypt (cost 14)** — Hash passwords, không lưu plaintext
- **API Key hashing** — Full key hash bằng SHA-256, chỉ lưu prefix cho display
- **Secure random** — `crypto/rand` cho token, key, password generation
- **HTTPS enforced** — Traefik auto TLS via Let's Encrypt

### A03 — Injection

- **Parameterized queries** — GORM parameterized queries, không raw SQL concatenation
- **Input validation** — Validate ở handler layer trước khi đưa vào database
- **Docker command safety** — Sử dụng Docker SDK (Go client), không shell exec

### A04 — Insecure Design

- **Defense in depth** — Multiple layers: Traefik → Rate limiter → Auth middleware → Handler validation
- **Principle of least privilege** — API keys chỉ có scopes cần thiết
- **Fail-safe defaults** — Mọi endpoint require auth trừ `/health`, `/api/auth/login`, `/api/auth/register`

### A05 — Security Misconfiguration

- **Security headers** (via middleware):

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

- **CORS configurable** — Không wildcard `*` trong production
- **Docker socket protection** — Chỉ My PaaS server access Docker socket

### A06 — Vulnerable Components

- **Minimal dependencies** — Backend chỉ dùng: Fiber, GORM, jwt-go, bcrypt, uuid, godotenv
- **No ORM magic** — Simple GORM usage, explicit queries
- **Alpine base images** — Minimal attack surface cho containers

### A07 — Authentication Failures

- **Triple auth model**:
  1. JWT (access + refresh tokens)
  2. API Keys (prefix `mpk_live_` / `mpk_test_`)
  3. Session-based (optional)
- **Rate limiting on auth** — Limit login attempts
- **Token refresh** — Short-lived access token, long-lived refresh token
- **Secure logout** — Invalidate session/token

### A08 — Data Integrity

- **Deployment verification** — Health check sau deploy, failed = rollback
- **Audit trail** — Ghi log mọi thao tác thay đổi dữ liệu
- **Webhook signature** — Verify GitHub webhook signatures

### A09 — Logging & Monitoring

- **Structured logging** — Request logger với method, path, status, latency
- **Audit log model** — `audit_logs` table: user, action, resource, details, IP, timestamp
- **Prometheus metrics** — Request count, latency histogram, active deployments, error rates
- **Notification alerts** — Deploy fail → webhook notification

### A10 — Server-Side Request Forgery (SSRF)

- **Git URL validation** — Chỉ accept HTTP(S) git URLs
- **Internal network protection** — Services chạy trong Docker overlay network, không expose ra host mặc định

## Rate Limiting

| Mode | Backend | Shared across instances |
|---|---|---|
| Redis | Redis `INCR` + `EXPIRE` | ✅ |
| In-memory | `sync.Map` + goroutine cleanup | ❌ |

Default: **60 requests/minute** per IP.

## Encryption at Rest

```
Secret ENV var
    ↓
AES-256-GCM encrypt (ENCRYPTION_KEY from env)
    ↓
Base64 encode
    ↓
Store in database
    ↓
Decrypt only when injecting into container
```

::: danger
`ENCRYPTION_KEY` phải được bảo vệ cẩn thận. Mất key = mất toàn bộ secret values.
:::

## Network Security

```
Internet → Traefik (TLS termination)
              ↓
         My PaaS Server (internal)
              ↓
         Docker Network (overlay, encrypted)
              ↓
         App Containers (isolated)
```

- Traefik xử lý TLS, chuyển traffic nội bộ qua HTTP
- Containers không expose port ra host trừ khi cấu hình explicit
- Docker overlay network encrypt traffic giữa nodes (optional `--opt encrypted`)
