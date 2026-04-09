# Roadmap

Kế hoạch phát triển My PaaS trong tương lai, được sắp xếp theo mức độ ưu tiên.

## v4.1 — Monitoring & Observability

**Mục tiêu**: Dashboard giám sát nâng cao

- [ ] **Grafana integration** — Auto-provision Grafana dashboards cho mỗi project
- [ ] **Alert rules** — Cấu hình alert khi CPU > 80%, Memory > 90%, container restart
- [ ] **Log aggregation** — Tập trung logs từ tất cả containers vào Loki/Elasticsearch
- [ ] **Uptime monitoring** — HTTP health check endpoint monitoring với response time tracking
- [ ] **Cost estimation** — Ước tính chi phí resource usage per project

## v4.2 — CI/CD Pipeline

**Mục tiêu**: Build pipeline tích hợp

- [ ] **Build logs streaming** — Real-time build output qua WebSocket
- [ ] **Build cache** — Docker layer cache giữa các builds
- [ ] **Multi-stage preview** — Deploy preview environments cho pull requests
- [ ] **Pipeline steps** — Configurable build → test → deploy pipeline
- [ ] **GitHub/GitLab integration** — Commit status checks, PR comments

## v4.3 — Horizontal Scaling

**Mục tiêu**: Auto-scaling và multi-node

- [ ] **Auto-scaling rules** — Scale dựa trên CPU/Memory/Request count
- [ ] **Load balancer health checks** — Active health checking cho backend services
- [ ] **Blue/Green deployment** — Zero-downtime deployment strategy
- [ ] **Canary releases** — Gradual rollout (10% → 50% → 100%)
- [ ] **Sticky sessions** — Session affinity cho stateful apps

## v4.4 — Developer Experience

**Mục tiêu**: Tối ưu workflow cho developers

- [ ] **Web terminal** — SSH into containers từ dashboard
- [ ] **Remote debugging** — Attach debugger vào running container
- [ ] **File manager** — Browse/edit files trong container volumes
- [ ] **Database GUI** — Adminer/pgAdmin integration cho linked services
- [ ] **CLI auto-update** — Self-update mechanism cho mypaas-cli
- [ ] **VS Code Extension v2** — Deploy trực tiếp từ VS Code, live logs panel

## v4.5 — Marketplace Expansion

**Mục tiêu**: One-click stacks phong phú

- [ ] **Community templates** — User-contributed marketplace templates
- [ ] **Stack bundles** — Full-stack templates (React + API + DB + Cache)
- [ ] **Plugin system** — Extensible providers cho custom languages/frameworks
- [ ] **Buildpack support** — Cloud Native Buildpack compatibility
- [ ] **Helm chart import** — Import Kubernetes Helm charts as templates

## v5.0 — Kubernetes Support

**Mục tiêu**: Hỗ trợ Kubernetes orchestrator

- [ ] **Dual orchestrator** — Swarm hoặc Kubernetes backend
- [ ] **Kubernetes provider** — Deploy as Deployment + Service + Ingress
- [ ] **Namespace isolation** — Mỗi org/user một namespace
- [ ] **Resource quotas** — Kubernetes ResourceQuota per namespace
- [ ] **Service mesh** — Istio/Linkerd integration cho observability

## v5.1 — Multi-Cloud

**Mục tiêu**: Deploy trên nhiều cloud provider

- [ ] **AWS integration** — ECS/EKS deployment target
- [ ] **GCP integration** — Cloud Run/GKE deployment target
- [ ] **Azure integration** — Azure Container Apps deployment target
- [ ] **Hybrid cloud** — Cross-cloud service mesh
- [ ] **Edge deployment** — Deploy lightweight containers tới edge nodes

## Nguyên tắc phát triển

Khi phát triển các tính năng mới, My PaaS tuân thủ:

1. **Backward compatible** — Không break API/CLI hiện tại
2. **Optional complexity** — Tính năng nâng cao là optional, defaults luôn đơn giản
3. **Self-hosted first** — Mọi tính năng chạy được trên single VPS
4. **No vendor lock-in** — Không phụ thuộc cloud provider cụ thể
5. **Security by default** — Mọi tính năng mới phải qua security review

## Đóng góp

Muốn đóng góp? Xem [GitHub Issues](https://github.com/your-org/my-paas/issues) hoặc tạo Feature Request.
