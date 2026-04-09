# Tại sao My PaaS?

## Bối cảnh thị trường

Hệ sinh thái PaaS hiện tại có thể chia thành 3 nhóm:

| Nhóm | Đại diện | Ưu điểm | Nhược điểm |
|---|---|---|---|
| **PaaS thương mại** | Heroku, Railway, Render, Fly.io | UX tốt, zero-ops | Chi phí cao, vendor lock-in |
| **Self-hosted phức tạp** | Kubernetes, Nomad, OpenShift | Linh hoạt tối đa | Đường cong học tập dốc, cần đội ops |
| **Self-hosted đơn giản** | Dokku, CapRover, Coolify | Chi phí thấp, kiểm soát dữ liệu | Thiếu enterprise features, UX hạn chế |

**My PaaS** thuộc nhóm thứ 3 nhưng hướng đến thu hẹp khoảng cách với nhóm 1 về UX và nhóm 2 về tính năng enterprise.

## Vấn đề thực tế mà My PaaS giải quyết

### 1. "Deploy cái app Node.js lên server mất 2 tiếng"

Một developer mới muốn đưa app lên VPS phải:
- SSH vào server, cài Node, cài Nginx, viết config
- Tạo systemd service, cấu hình SSL
- Mỗi lần update: `git pull && npm build && pm2 restart`

**Với My PaaS:** Push code → auto-detect Node.js → build Docker image → deploy với health check → auto-SSL qua Traefik. Toàn bộ mất ~2 phút.

### 2. "Staging server ai cũng sợ đụng vào"

Môi trường staging dùng chung, không có isolation:
- App A crash → ảnh hưởng App B
- Không biết ai deploy lần cuối, deploy cái gì
- Rollback = "anh ơi backup đâu rồi?"

**Với My PaaS:** Mỗi app là một Docker container/service riêng biệt. Audit log ghi nhận mọi thao tác. Rollback 1 click về deployment trước đó.

### 3. "Kubernetes cho team 3 người thì overkill quá"

K8s giải quyết được mọi thứ nhưng:
- Cần ít nhất 3 node cho HA control plane
- Developer phải học YAML manifests, Helm charts
- Ops overhead lớn: etcd backup, cert rotation, network policies

**Với My PaaS + Docker Swarm:** 2 node là đủ cho HA. Không cần manifest file. Quản lý qua web UI hoặc CLI.

### 4. "Heroku tính $7/dyno, 10 app = $70/tháng"

Với startup giai đoạn đầu, $70/tháng cho staging + production là đáng kể.

**Với My PaaS:** VPS $5–10/tháng (4GB RAM) chạy được 10–20 app nhẹ cùng lúc. Database, Redis, reverse proxy tất cả trên cùng một node.

## Quyết định thiết kế then chốt

### Tại sao single binary?

My PaaS server là **một Go binary duy nhất** phục vụ cả API lẫn frontend:

- **Đơn giản hoá deployment** — `COPY --from=builder /mypaas-server .` là xong
- **Không cần orchestrate nhiều service** — không có API gateway, frontend server riêng
- **Cold start nhanh** — Go binary khởi động trong ~100ms
- **Resource footprint nhỏ** — ~30MB RAM idle

### Tại sao Docker Swarm thay vì Kubernetes?

| Tiêu chí | Docker Swarm | Kubernetes |
|---|---|---|
| Setup | `docker swarm init` (1 lệnh) | kubeadm/k3s/minikube (phức tạp) |
| Tài nguyên tối thiểu | 1 node, ~50MB | 1 node, ~500MB+ |
| Learning curve | Docker CLI quen thuộc | Hoàn toàn mới (kubectl, manifests) |
| Rolling updates | Built-in | Built-in nhưng config phức tạp hơn |
| Đối tượng | Small–medium teams | Medium–large teams |

Swarm là lựa chọn phù hợp cho đối tượng mục tiêu: **developer cá nhân và team nhỏ** cần orchestration cơ bản mà không cần cluster management phức tạp.

### Tại sao SQLite làm default database?

- **Zero dependency** — không cần PostgreSQL container khi mới bắt đầu
- **Single-file backup** — `cp mypaas.db mypaas.db.bak`
- **WAL mode** — hỗ trợ concurrent reads tốt
- **Chuyển đổi dễ** — khi cần scale, switch sang PostgreSQL bằng 1 biến môi trường

## Hành trình phát triển

My PaaS trải qua 4 giai đoạn iterative:

```
v1 (IDEA)          v2 (Extension)        v3 (Docker Service)    v4 (Enterprise)
┌──────────┐       ┌──────────────┐       ┌────────────────┐     ┌─────────────────┐
│ Microsvcs│  ───► │ VS Code drag │  ───► │ Single binary  │ ──► │ JWT, Orgs,      │
│ Railway  │       │ & drop Docker│       │ Docker socket  │     │ PostgreSQL,     │
│ clone    │       │ designer     │       │ Git watcher    │     │ Redis, Metrics  │
└──────────┘       └──────────────┘       └────────────────┘     └─────────────────┘
    ✗ Quá phức tạp     ✓ Giữ lại extension    ✓ Core architecture    ✓ Production-ready
```

- **v1** — Thiết kế microservice ban đầu bị loại bỏ vì quá phức tạp cho mục tiêu ban đầu
- **v2** — VS Code Extension với Docker canvas designer — giữ lại như công cụ bổ trợ
- **v3** — Kiến trúc hiện tại: single binary + Docker socket mount
- **v4** — Enterprise features biến prototype thành production-grade platform
