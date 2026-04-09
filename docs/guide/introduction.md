# Giới thiệu My PaaS

**My PaaS** là một nền tảng triển khai ứng dụng mã nguồn mở (Platform as a Service), được xây dựng để giải quyết bài toán: *"Từ source code đến production nhanh nhất có thể, với chi phí vận hành thấp nhất."*

## Dự án ra đời từ đâu?

Trong thực tế phát triển phần mềm, việc triển khai một ứng dụng web lên server thường đòi hỏi:

1. **Viết Dockerfile thủ công** — mỗi dự án một cấu hình khác nhau
2. **Cấu hình CI/CD** — GitHub Actions, GitLab CI, Jenkins pipeline
3. **Quản lý reverse proxy** — Nginx/Traefik config, SSL certificates
4. **Quản lý database, cache** — PostgreSQL, Redis, volume mounts
5. **Giám sát & logging** — Prometheus, Grafana, ELK stack

Các nền tảng PaaS thương mại như Heroku, Railway, Render giải quyết tốt các vấn đề này, nhưng có hạn chế:

| Vấn đề | Chi tiết |
|---|---|
| **Chi phí cao** | $5–25/container/tháng, tăng nhanh khi scale |
| **Vendor lock-in** | Phụ thuộc vào hạ tầng và API riêng của nhà cung cấp |
| **Giới hạn tuỳ biến** | Không kiểm soát được network, storage, hoặc runtime config |
| **Dữ liệu ngoài tầm kiểm soát** | Data-residency regulations khó đáp ứng |

**My PaaS** ra đời để mang lại trải nghiệm PaaS đầy đủ trên chính hạ tầng của bạn — một VPS $5/tháng cũng đủ chạy.

## Tầm nhìn

> *Self-hosted PaaS cho developer — đơn giản như Heroku, linh hoạt như Kubernetes, nhẹ như Docker Compose.*

Dự án hướng đến:

- **Developer đơn lẻ** muốn deploy nhanh side project mà không cần học Kubernetes
- **Startup nhỏ** cần môi trường staging/production tự quản với chi phí thấp
- **Đội nhóm** cần multi-tenant platform với phân quyền, quotas, và audit trail
- **Doanh nghiệp** cần triển khai on-premise với yêu cầu bảo mật nghiêm ngặt

## Những gì đã thực hiện

My PaaS v4.0.0 hiện tại bao gồm:

### Platform Core
- **Auto-detection** 7 ngôn ngữ lập trình — tự sinh Dockerfile tối ưu
- **Git-based deployment** — clone → detect → build → deploy tự động
- **Docker Swarm orchestration** — multi-node cluster, rolling updates
- **Traefik reverse proxy** — auto-SSL, domain routing, load balancing
- **Backing services** — PostgreSQL, Redis, MySQL, MongoDB, MinIO one-click
- **Local Docker Registry** — lưu trữ image nội bộ

### Enterprise Features
- **Triple authentication** — JWT + API Key + Session Token
- **Multi-tenant organizations** — resource quotas, member management
- **AES-256-GCM encryption** — mã hoá secret tại chỗ (at-rest)
- **Prometheus metrics** — deployments, latency, queue depth
- **Notification channels** — Webhook, Slack với event-based rules
- **Audit logging** — ghi nhận mọi hành động admin

### Developer Tools
- **Web Dashboard** — React 19 SPA quản lý trực quan
- **mypaas-cli** — CLI tool đầy đủ cho terminal workflow
- **VS Code Extension** — kéo thả Docker service designer
- **OpenAPI/Swagger** — API documentation tương tác
- **Real-time SSE logs** — theo dõi build/runtime logs trực tiếp

## Khởi chạy nhanh

```bash
# Clone repository
git clone https://github.com/my-paas/my-paas.git
cd my-paas

# Khởi chạy với Docker Compose
docker compose -f deploy/docker-compose.yml up -d

# Truy cập dashboard
open http://localhost:8080
```

Lần đầu truy cập, hệ thống sẽ yêu cầu tạo tài khoản admin (setup mode).

## Kiến trúc tổng quan

```
┌────────────────────────────────────────────────┐
│                   Internet                      │
└──────────────────┬─────────────────────────────┘
                   │
┌──────────────────▼─────────────────────────────┐
│           Traefik (port 80/443)                 │
│         Auto-SSL · Domain routing               │
└──────────────────┬─────────────────────────────┘
                   │
┌──────────────────▼─────────────────────────────┐
│         My PaaS Server (Go + Fiber)             │
│  ┌─────────┐ ┌──────────┐ ┌─────────────────┐  │
│  │ REST API│ │ React SPA│ │ Deploy Workers  │  │
│  └────┬────┘ └──────────┘ └───────┬─────────┘  │
│       │                           │             │
│  ┌────▼────────────────┐   ┌──────▼──────────┐  │
│  │  SQLite/PostgreSQL  │   │  Docker Engine  │  │
│  └─────────────────────┘   └───────┬─────────┘  │
└────────────────────────────────────┤────────────┘
                                     │
┌────────────────────────────────────▼────────────┐
│              mypaas-network (overlay)            │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ App #1  │ │ App #2  │ │ DB/Redis│  ...       │
│  └─────────┘ └─────────┘ └─────────┘           │
└─────────────────────────────────────────────────┘
```

Xem chi tiết tại [Tổng quan kiến trúc](/architecture/overview).
