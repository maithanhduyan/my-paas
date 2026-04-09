# Tiêu chuẩn chất lượng

## Nguyên tắc thiết kế

### 1. Simplicity First

- **Single binary** — Toàn bộ backend + static files đóng gói trong 1 file Go binary (~20MB)
- **Zero external runtime** — Không cần Java, Node.js, Python trên server
- **SQLite by default** — Chạy ngay không cần cài database
- **Sane defaults** — Mọi cấu hình đều có giá trị mặc định hợp lý

### 2. Convention over Configuration

- Detect ngôn ngữ tự động → không cần viết Dockerfile thủ công
- Tự tạo `BuildPlan` → Dockerfile → Docker image → Deploy
- Default network, default domain routing — developer chỉ cần push code

### 3. Graceful Degradation

| Feature | Có Redis | Không Redis |
|---|---|---|
| Rate limiting | Redis-based, shared across instances | In-memory, per-instance |
| Queue | Redis list | In-memory channel |
| Cache | Redis cache | In-memory map |
| Session | Redis store | Memory store |

Hệ thống **luôn hoạt động** dù thiếu optional dependencies.

## Tiêu chuẩn code

### Go Backend

- **Error handling**: Mọi error đều được wrap và trả về HTTP response phù hợp
- **Middleware chain**: CORS → Logger → Rate Limiter → Auth → Handler
- **Data validation**: Validate ở handler layer trước khi xử lý
- **Secret management**: Mọi giá trị nhạy cảm mã hoá AES-256-GCM tại rest
- **Database migrations**: Auto-migrate schema khi khởi động (`store.AutoMigrate()`)
- **Structured responses**: Tất cả API trả JSON format nhất quán

```go
// Consistent error response
c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
    "error": "validation failed",
    "details": "project name is required",
})
```

### Frontend (React)

- **TypeScript strict** — Không `any` type
- **Component-based** — Mỗi page/feature là một component riêng
- **State management** — Zustand cho global state, React hooks cho local state
- **Responsive** — Tailwind CSS responsive-first design
- **Error boundaries** — Bắt lỗi render, hiển thị fallback UI

### Infrastructure

- **Health checks** — `GET /health` trả status + Docker connection + version
- **Container isolation** — Mỗi app chạy trong Docker container/service riêng
- **Network segmentation** — `mypaas-network` overlay network
- **Volume persistence** — Named volumes cho data services
- **Log rotation** — Docker log driver giới hạn size

## Deployment Standards

### Build pipeline đảm bảo

| Bước | Kiểm tra |
|---|---|
| Clone | Git URL hợp lệ, branch tồn tại |
| Detect | Nhận diện được ngôn ngữ (có ít nhất 1 provider match) |
| Build | Docker build thành công, image có size hợp lý |
| Deploy | Container/service start, port mở |
| Health | Container healthy trong 30s, không crash loop |

### Rollback tự động

Nếu deployment thất bại ở bất kỳ bước nào:
1. Status chuyển sang `failed`
2. Container cũ vẫn chạy (không bị stop trước khi container mới healthy)
3. Có thể rollback thủ công về deployment trước: `POST /deployments/:id/rollback`

### Resource management

```yaml
# Default limits
CPU:    1.0 core
Memory: 512MB
# Configurable per project via API
```

## Performance

### Response time targets

| Endpoint type | Target |
|---|---|
| Health check | < 10ms |
| CRUD operations | < 50ms |
| List with pagination | < 100ms |
| Deploy trigger | < 500ms (async) |
| Log streaming | Real-time (SSE) |

### Scalability

| Component | Standalone | Swarm (3 nodes) |
|---|---|---|
| Max concurrent apps | ~50 | ~200+ |
| Max services | ~20 | ~100+ |
| Deploy throughput | 3 concurrent | 9 concurrent (3 workers × 3 nodes) |
| Database | SQLite (single writer) | PostgreSQL (connection pooling) |

## Monitoring

- **Prometheus metrics** — `/api/metrics` endpoint với 10+ custom metrics
- **Container stats** — CPU%, Memory, Network I/O per container
- **Deployment tracking** — Lịch sử deploy, duration, status
- **Audit logging** — Ghi lại mọi thao tác quan trọng (tạo/xoá project, deploy, thay đổi env)
- **Notification system** — Alert qua webhook/email khi deploy fail/succeed
