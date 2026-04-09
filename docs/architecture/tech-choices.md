# Lựa chọn công nghệ

Mỗi quyết định công nghệ trong My PaaS đều được cân nhắc dựa trên 3 tiêu chí:

1. **Phù hợp mục tiêu** — self-hosted, nhẹ, dễ vận hành
2. **Đáp ứng nhu cầu thực tế** — giải quyết bài toán triển khai ứng dụng
3. **Chi phí bảo trì thấp** — ít dependency, ít complexity

## Backend: Go + Fiber

### Tại sao Go?

| Tiêu chí | Go | Node.js | Python | Rust |
|---|---|---|---|---|
| Compile thành binary | ✅ Single binary | ❌ Runtime cần | ❌ Runtime cần | ✅ Single binary |
| Memory footprint | ~30 MB | ~80 MB | ~60 MB | ~10 MB |
| Concurrency model | Goroutines (M:N) | Event loop (single-thread) | Threading/async | Tokio (M:N) |
| Docker SDK | Chính thức | Community | Community | Community |
| Build time | ~60s | N/A | N/A | ~300s |
| Learning curve | Thấp | Thấp | Thấp | Cao |

**Go được chọn vì:**

- **Single binary deployment** — copy 1 file vào Alpine container là xong
- **Docker SDK chính thức** — `github.com/docker/docker` được maintain bởi chính Docker team
- **Goroutine** — deploy workers, git watcher, SSE streaming chạy concurrent tự nhiên
- **CGO cho SQLite** — `go-sqlite3` hoạt động tốt, WAL mode ổn định
- **Startup time** — ~100ms cold start, quan trọng khi container restart

### Tại sao Fiber thay vì Gin/Echo/Chi?

| Framework | RPS (benchmark) | API style | Middleware | Fiber-like? |
|---|---|---|---|---|
| **Fiber** | ~300K | Express-like | Stack-based | — |
| Gin | ~250K | Similar | Similar | Gần |
| Echo | ~200K | Similar | Similar | Gần |
| Chi | ~150K | net/http compatible | Standard | Khác |

Fiber được chọn vì:
- **Fasthttp engine** — nhanh hơn `net/http` 2–3x cho I/O bound
- **Express.js-like API** — quen thuộc với developer đến từ Node.js
- **Built-in SSE support** — quan trọng cho real-time log streaming
- **Body parsing, validation** — giảm boilerplate
- **Active maintenance** — 30K+ stars, release thường xuyên

## Frontend: React 19 + Vite + Tailwind

### Tại sao React?

- **Ecosystem lớn nhất** — dễ tìm developer, nhiều thư viện hỗ trợ
- **Component model** — phù hợp cho dashboard UI phức tạp
- **React 19** — Server Components ready, concurrent features

### Tại sao Vite thay vì Webpack/Turbopack?

- **Dev server nhanh** — ESM-based, HMR trong ~50ms
- **Build nhanh** — Rollup-based production build
- **Zero config** — React plugin là đủ
- **Nhẹ** — `node_modules` nhỏ hơn Webpack đáng kể

### Tại sao Tailwind CSS?

- **Utility-first** — không cần viết CSS files riêng
- **Purge built-in** — CSS bundle siêu nhỏ (~10KB gzipped)
- **Responsive** — mobile-first classes sẵn có
- **Consistency** — design tokens (spacing, colors) nhất quán

### Tại sao Zustand thay vì Redux?

| Tiêu chí | Zustand | Redux Toolkit |
|---|---|---|
| Bundle size | ~1 KB | ~13 KB |
| Boilerplate | Minimal | Moderate (slices, reducers) |
| API | Hook-based | Hook + Provider |
| Middleware | Simple | Redux middleware chain |
| Learning curve | 5 phút | 1–2 giờ |

Dashboard PaaS không cần time-travel debugging hay complex state machines → Zustand đủ dùng.

## Database: SQLite → PostgreSQL

### Dual-driver strategy

```
Developer/Staging         →  SQLite (zero-config)
Production/Enterprise     →  PostgreSQL (concurrent, backup)
```

**SQLite** là default vì:
- Không cần container database riêng
- Single-file backup: `cp mypaas.db backup.db`
- WAL mode cho phép concurrent reads
- Phù hợp cho ≤50 concurrent users

**PostgreSQL** cho enterprise vì:
- ACID transactions với concurrent writes tốt hơn
- Connection pooling, streaming replication
- `pg_dump/pg_restore` cho backup chuyên nghiệp
- Hỗ trợ JSON operators, full-text search

Chuyển đổi bằng 1 biến: `MYPAAS_DB_DRIVER=postgres`.

## Orchestration: Docker Swarm

### So sánh với alternatives

| Feature | Docker Compose | Docker Swarm | Kubernetes |
|---|---|---|---|
| Multi-node | ❌ | ✅ | ✅ |
| Rolling updates | ❌ | ✅ | ✅ |
| Service discovery | Docker DNS | Overlay network DNS | CoreDNS |
| Load balancing | ❌ | Ingress routing mesh | Service/Ingress |
| Auto-healing | `restart: always` | Task rescheduling | Pod rescheduling |
| Min resources | ~0 MB overhead | ~50 MB | ~500 MB+ |
| Setup complexity | `docker compose up` | `docker swarm init` | kubeadm/k3s (nhiều bước) |

Swarm là **sweet spot** giữa simplicity và capability cho target audience.

## Security Stack

| Layer | Công nghệ | Lý do |
|---|---|---|
| **Password hashing** | bcrypt (`golang.org/x/crypto`) | Industry standard, adaptive cost |
| **JWT signing** | HMAC-SHA256 (custom) | Không cần external lib, stateless auth |
| **Secret encryption** | AES-256-GCM (custom) | Authenticated encryption, Go stdlib `crypto/aes` |
| **Rate limiting** | Redis BRPOPLPUSH / in-memory sliding window | Distributed khi scale, graceful fallback |
| **Transport security** | Traefik + Let's Encrypt ACME | Auto-renewal, zero-config SSL |

## Reverse Proxy: Traefik v2

### Tại sao Traefik thay vì Nginx?

| Feature | Traefik | Nginx |
|---|---|---|
| Docker-native | ✅ Label-based auto-config | ❌ Config file cần reload |
| Let's Encrypt | ✅ Built-in ACME | ❌ Cần certbot + cron |
| Swarm support | ✅ swarmMode=true | ❌ Cần extra setup |
| Dashboard | ✅ Built-in UI | ❌ Cần thêm |
| Hot reload | ✅ Tự động | ❌ `nginx -s reload` |

Traefik phù hợp hoàn hảo cho Docker-first architecture — mỗi app chỉ cần Docker label, không cần config file.

## Monitoring: Prometheus

Chọn Prometheus metrics format vì:
- **De facto standard** — mọi monitoring tool đều hỗ trợ
- **Text-based** — dễ debug bằng `curl`
- **Pull model** — Prometheus scrape endpoint, không cần push agent
- **Custom metrics** — Go stdlib đủ để format, không cần client library

## Dependencies tối thiểu

My PaaS cố gắng **giữ dependency tree nhỏ**:

### Go (9 direct dependencies)

| Package | Mục đích |
|---|---|
| `docker/docker` | Docker Engine SDK |
| `docker/go-connections` | Port mapping utilities |
| `gofiber/fiber/v2` | HTTP framework |
| `google/uuid` | ID generation |
| `lib/pq` | PostgreSQL driver |
| `mattn/go-sqlite3` | SQLite driver |
| `moby/go-archive` | Tar archiving |
| `redis/go-redis/v9` | Redis client |
| `golang.org/x/crypto` | bcrypt hashing |

Không dùng ORM (raw SQL), không dùng DI framework, không dùng external JWT/encryption library.

### Frontend (5 runtime dependencies)

| Package | Mục đích |
|---|---|
| `react` + `react-dom` | UI framework |
| `react-router-dom` | Routing |
| `zustand` | State management |
| `lucide-react` | Icons |
