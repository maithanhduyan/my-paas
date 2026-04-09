# Cấu hình

My PaaS sử dụng biến môi trường để cấu hình. Không cần file config phức tạp.

## Biến môi trường

### Core

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_LISTEN` | `:8080` | Địa chỉ lắng nghe |
| `MYPAAS_DOMAIN` | `localhost` | Domain gốc cho subdomain routing |
| `MYPAAS_DATA_DIR` | `/data` | Thư mục lưu trữ dữ liệu |
| `MYPAAS_BUILDS_DIR` | `/data/builds` | Thư mục chứa build context |

### Database

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_DB_DRIVER` | `sqlite` | `sqlite` hoặc `postgres` |
| `MYPAAS_DB` | `/data/mypaas.db` | Đường dẫn file SQLite |
| `MYPAAS_DB_URL` | — | PostgreSQL connection string |

### Authentication & Security

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_SECRET` | auto-generated | JWT signing key (≥32 bytes) |
| `MYPAAS_ENCRYPTION_KEY` | auto-generated | AES-256-GCM key (64 hex chars) |
| `MYPAAS_JWT_EXPIRY` | `24h` | Access token TTL |
| `MYPAAS_REFRESH_EXPIRY` | `720h` | Refresh token TTL (30 ngày) |
| `MYPAAS_REGISTRATION_OPEN` | `false` | Cho phép đăng ký công khai |

### Redis (tuỳ chọn)

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_REDIS_URL` | — | URL Redis (ví dụ: `redis://redis:6379`) |

Khi bật Redis, hệ thống sử dụng cho: rate limiting, job queue, cache, pub/sub. Khi không có Redis, tự động fallback về in-memory.

### Rate Limiting

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_RATE_LIMIT_RPS` | `100` | Request/giây/IP |
| `MYPAAS_RATE_LIMIT_BURST` | `200` | Burst limit |

### Workers

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_WORKER_COUNT` | `2` | Số deploy worker chạy song song |
| `MYPAAS_QUEUE_SIZE` | `100` | Kích thước hàng đợi job |

### Operations

| Biến | Mặc định | Mô tả |
|---|---|---|
| `MYPAAS_MAINTENANCE_MODE` | `false` | Bật chế độ bảo trì (trả 503 cho authenticated requests) |
| `ACME_EMAIL` | `admin@example.com` | Email cho Let's Encrypt certificates |

## Ví dụ cấu hình

### Development (tối thiểu)

```env
MYPAAS_LISTEN=:8080
```

### Production (standalone)

```env
MYPAAS_LISTEN=:8080
MYPAAS_DOMAIN=paas.company.com
MYPAAS_SECRET=a-very-long-random-secret-at-least-32-bytes
MYPAAS_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
ACME_EMAIL=devops@company.com
```

### Enterprise (full stack)

```env
MYPAAS_LISTEN=:8080
MYPAAS_DOMAIN=paas.company.com
MYPAAS_DB_DRIVER=postgres
MYPAAS_DB_URL=postgres://mypaas:S3cret!@postgres:5432/mypaas?sslmode=disable
MYPAAS_REDIS_URL=redis://redis:6379
MYPAAS_SECRET=enterprise-jwt-secret-minimum-32-bytes-long
MYPAAS_ENCRYPTION_KEY=aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899
MYPAAS_WORKER_COUNT=4
MYPAAS_QUEUE_SIZE=200
MYPAAS_RATE_LIMIT_RPS=200
MYPAAS_RATE_LIMIT_BURST=500
ACME_EMAIL=devops@company.com
```
