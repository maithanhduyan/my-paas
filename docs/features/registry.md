# Docker Registry

## Tổng quan

My PaaS tích hợp **local Docker Registry** (registry:2) để lưu trữ image nội bộ, không phụ thuộc Docker Hub hay registry bên ngoài.

## Use cases

- **Air-gapped deployment** — triển khai trong mạng nội bộ không có internet
- **Faster pulls** — worker nodes pull image từ local registry thay vì rebuild
- **Image versioning** — lưu trữ mọi phiên bản đã build
- **Cross-node sharing** — Swarm worker nodes pull từ registry thay vì transfer tar

## API

### Kiểm tra trạng thái

```bash
GET /api/registry/status
# → { "status": "running", "address": "localhost", "port": "5000" }
# → { "status": "not_running" }
```

### Khởi động registry

```bash
POST /api/registry/start    # (admin only)
# → { "message": "registry started", "address": "localhost:5000" }
```

Hệ thống tự động:
- Deploy `registry:2` image dưới dạng Swarm service (hoặc container)
- Publish port 5000
- Mount volume `mypaas-registry-data` cho persistent storage
- Enable delete API (`REGISTRY_STORAGE_DELETE_ENABLED=true`)

### Dừng registry

```bash
POST /api/registry/stop    # (admin only)
```

### Push project image

```bash
POST /api/projects/:id/push
# → { "message": "pushed to registry", "image": "localhost:5000/my-app:latest" }
```

Thực hiện:
1. Tag image `mypaas-<name>:latest` → `localhost:5000/<name>:latest`
2. `docker push` tới local registry

### Liệt kê images

```bash
GET /api/registry/images
# → [{ "name": "my-app", "tags": ["latest", "abc123"] }]
```

### Xoá image

```bash
DELETE /api/registry/images/:name?tag=latest    # (admin only)
```

## Cấu hình Docker Daemon

Để push/pull từ local registry (HTTP, không HTTPS), Docker daemon cần cấu hình insecure registry:

```json
// /etc/docker/daemon.json
{
  "insecure-registries": [
    "mypaas-registry:5000",
    "127.0.0.1:5000",
    "localhost:5000"
  ]
}
```

Sau khi thay đổi: `systemctl restart docker`

::: warning
Cấu hình này cần áp dụng trên **tất cả nodes** trong Swarm cluster.
:::
