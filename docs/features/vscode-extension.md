# VS Code Extension

## Tổng quan

**My PaaS — Docker Visual Designer** là VS Code extension cho phép kéo thả thiết kế Docker service stack trực quan, tự sinh `docker-compose.yml`, và quản lý container lifecycle ngay trong editor.

## Tính năng

### Canvas Designer

- **Drag & drop** các service blocks lên canvas
- **Visual connections** giữa services (network links)
- **Real-time preview** docker-compose.yml khi thiết kế
- **Export** docker-compose.yml vào workspace

### Service Templates

Templates định sẵn cho các service phổ biến:

- **Databases**: PostgreSQL, MySQL, MongoDB, Redis
- **Web servers**: Nginx, Apache, Traefik
- **Applications**: Node.js, Python, Go, PHP
- **Tools**: Adminer, phpMyAdmin, MinIO

### Container Management

- **Compose Up/Down** — khởi động/dừng docker-compose stack
- **Container logs** — xem logs trực tiếp trong VS Code
- **Container stats** — CPU, memory, network

### Auto Detection

- **Detect ngôn ngữ** project hiện tại (sử dụng core module)
- **Suggest Dockerfile** tối ưu cho project
- **Suggest services** phụ thuộc (ví dụ: Django → suggest PostgreSQL)

## Cài đặt

```bash
cd my-paas-extension
bun install
bun run build

# Package VSIX
bun run package
# → my-paas-extension-<version>.vsix
```

## Kiến trúc extension

```
my-paas-extension/
├── src/
│   ├── extension.ts          # Entry point, register commands
│   ├── core/
│   │   └── CoreBridge.ts     # Bridge to Go core binary
│   ├── docker/
│   │   ├── ComposeGenerator.ts   # Generate docker-compose.yml
│   │   └── DockerManager.ts      # Container lifecycle
│   ├── panels/
│   │   └── CanvasPanel.ts        # Webview panel management
│   ├── project/
│   │   └── ProjectManager.ts     # Project detection
│   ├── sidebar/
│   │   └── SidebarProviders.ts   # Activity bar views
│   └── templates/
│       └── TemplateRegistry.ts   # Service template registry
└── webview/
    └── src/
        ├── App.tsx               # React canvas UI
        ├── components/           # Canvas components
        ├── hooks/                # Custom hooks
        └── store/                # Zustand state
```

Extension sử dụng **webview** (React + Vite) để render canvas UI, giao tiếp với extension host qua `postMessage()`.

## Commands

| Command | Mô tả |
|---|---|
| `My PaaS: Open Canvas` | Mở canvas designer |
| `My PaaS: Compose Up` | Chạy docker-compose up |
| `My PaaS: Compose Down` | Dừng docker-compose |
| `My PaaS: Generate Compose` | Sinh docker-compose.yml từ canvas |
| `My PaaS: Auto Detect` | Detect ngôn ngữ project |

## Cross-platform

Extension đóng gói core Go binary cho từng platform:

- `win32-x64`, `win32-arm64`
- `linux-x64`, `linux-arm64`
- `darwin-x64`, `darwin-arm64`
