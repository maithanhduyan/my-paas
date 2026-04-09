---
layout: home

hero:
  name: My PaaS
  text: Nền tảng triển khai ứng dụng tự quản
  tagline: Từ Git push đến production — trong một lệnh duy nhất
  actions:
    - theme: brand
      text: Bắt đầu
      link: /guide/introduction
    - theme: alt
      text: Xem Roadmap
      link: /roadmap

features:
  - icon: 🚀
    title: Zero-config deployment
    details: Tự động nhận diện 7 ngôn ngữ (Go, Node, Python, Java, PHP, Rust, Static) — sinh Dockerfile và triển khai không cần cấu hình thủ công.
  - icon: 🐳
    title: Docker Swarm native
    details: Hỗ trợ cả standalone container và Docker Swarm cluster. Rolling update, health check, auto-rollback tích hợp sẵn.
  - icon: 🔒
    title: Enterprise-grade security
    details: JWT + API Key + Session triple-auth, AES-256-GCM mã hoá bí mật, rate limiting, security headers theo OWASP Top 10.
  - icon: 📊
    title: Observability
    details: Prometheus metrics, real-time SSE logs, notification webhooks/Slack, audit trail — giám sát toàn diện mọi hành động.
  - icon: 🏢
    title: Multi-tenant
    details: Tổ chức (Organization) với phân quyền, resource quotas, API key scoping — sẵn sàng cho đội nhóm.
  - icon: 🛠️
    title: CLI + VS Code Extension
    details: Quản lý từ terminal với mypaas-cli hoặc kéo thả trực quan qua VS Code Extension.
---
