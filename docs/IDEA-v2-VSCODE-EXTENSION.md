# My PaaS — VS Code Extension với giao diện kéo thả Docker

## Tổng quan

Thay vì build một full PaaS platform (frontend riêng, backend riêng, worker riêng...), ta đơn giản hóa thành **một VS Code Extension** duy nhất:

- Giao diện **kéo thả (drag & drop)** để thiết kế hạ tầng Docker
- Mỗi service là một **block/node** trên canvas (PostgreSQL, Redis, Node.js app, Go app, Nginx...)
- Kéo đường nối giữa các block để định nghĩa **network connections**
- Tự động sinh **docker-compose.yml**
- Quản lý container trực tiếp từ VS Code (start, stop, restart, logs, shell)
- Không cần dashboard web riêng, không cần deploy backend

---

## Tại sao VS Code Extension?

| Full PaaS (Railway clone) | VS Code Extension |
|---|---|
| Cần frontend + backend + worker + DB + queue | Chỉ 1 extension |
| Phải deploy platform riêng | Chạy local, không deploy gì |
| Mất hàng tháng build | MVP trong 2-3 tuần |
| Cần server riêng chạy platform | Chỉ cần máy có Docker |
| Phức tạp, nhiều moving parts | Đơn giản, tập trung |

---

## Kiến trúc Extension

```
┌─────────────────────────────────────────────┐
│              VS Code Extension              │
│                                             │
│  ┌─────────────────────────────────────┐    │
│  │     Webview Panel (React Canvas)    │    │
│  │                                     │    │
│  │  ┌─────┐    ┌─────┐    ┌─────┐     │    │
│  │  │ DB  │───▶│ API │───▶│Nginx│     │    │
│  │  └─────┘    └─────┘    └─────┘     │    │
│  │  ┌─────┐    ┌─────┐               │    │
│  │  │Redis│───▶│Worker│               │    │
│  │  └─────┘    └─────┘               │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  ┌─────────────────────────────────────┐    │
│  │       Extension Backend (TS)        │    │
│  │                                     │    │
│  │  • Docker Engine API (Dockerode)    │    │
│  │  • docker-compose generate/run      │    │
│  │  • Container lifecycle management   │    │
│  │  • Log streaming                    │    │
│  │  • File system (save/load projects) │    │
│  └─────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────┐
│   Docker Engine      │
│   (local machine)    │
└─────────────────────┘
```

---

## Thành phần chính

### 1. Canvas kéo thả (Webview)

**Tech stack:**
- React (trong VS Code Webview)
- React Flow (thư viện kéo thả node/edge tốt nhất)
- Tailwind CSS
- VS Code Webview UI Toolkit

**Chức năng:**
- Sidebar chứa danh sách **service templates** (kéo ra canvas):
  - Databases: PostgreSQL, MySQL, MongoDB, Redis, MariaDB
  - Apps: Node.js, Go, Python, PHP, Rust, Static Site
  - Infrastructure: Nginx, Traefik, MinIO, RabbitMQ, Kafka
  - Tools: pgAdmin, Adminer, Grafana, Prometheus
- Mỗi node trên canvas hiển thị:
  - Tên service
  - Trạng thái (stopped / running / error)
  - Port mapping
  - Nút quick actions (start, stop, logs, terminal)
- Kéo đường nối giữa 2 node = tạo network link
- Click vào node = mở panel cấu hình (image, ports, env vars, volumes, depends_on)

### 2. Extension Backend (TypeScript)

**Chức năng chính:**

#### Docker Management
```typescript
// Sử dụng Dockerode để tương tác với Docker Engine
import Docker from 'dockerode';

// Quản lý container lifecycle
startContainer(serviceId)
stopContainer(serviceId)
restartContainer(serviceId)
removeContainer(serviceId)
getContainerLogs(serviceId)  // stream realtime
execInContainer(serviceId)   // attach terminal
getContainerStats(serviceId) // CPU, RAM, Network
```

#### Compose Generation
```typescript
// Từ canvas state → sinh docker-compose.yml
generateCompose(canvasState: CanvasState): string
// Chạy docker-compose up/down
composeUp(projectPath: string)
composeDown(projectPath: string)
```

#### Project Persistence
```typescript
// Lưu trạng thái canvas thành file JSON
// File: .mypaas/project.json
saveProject(canvasState: CanvasState)
loadProject(): CanvasState
```

### 3. Service Templates

Mỗi template là một JSON definition:

```json
{
  "id": "postgres",
  "name": "PostgreSQL",
  "icon": "database",
  "category": "database",
  "defaults": {
    "image": "postgres:16-alpine",
    "ports": ["5432:5432"],
    "environment": {
      "POSTGRES_USER": "postgres",
      "POSTGRES_PASSWORD": "postgres",
      "POSTGRES_DB": "mydb"
    },
    "volumes": ["postgres-data:/var/lib/postgresql/data"],
    "healthcheck": {
      "test": ["CMD-SHELL", "pg_isready -U postgres"],
      "interval": "10s",
      "timeout": "5s",
      "retries": 5
    }
  }
}
```

```json
{
  "id": "nodejs-app",
  "name": "Node.js App",
  "icon": "nodejs",
  "category": "app",
  "defaults": {
    "build": {
      "context": "./",
      "dockerfile": "Dockerfile"
    },
    "ports": ["3000:3000"],
    "environment": {
      "NODE_ENV": "production"
    },
    "depends_on": []
  }
}
```

Users có thể tạo **custom templates** và lưu trong `.mypaas/templates/`.

---

## File structure của Extension

```
my-paas-extension/
├── package.json                 # Extension manifest
├── tsconfig.json
├── src/
│   ├── extension.ts             # Entry point, register commands
│   ├── panels/
│   │   └── CanvasPanel.ts       # Webview panel manager
│   ├── docker/
│   │   ├── DockerManager.ts     # Dockerode wrapper
│   │   ├── ComposeGenerator.ts  # Canvas → docker-compose.yml
│   │   └── LogStreamer.ts       # Container log streaming
│   ├── templates/
│   │   ├── TemplateRegistry.ts  # Load & manage templates
│   │   └── defaults/            # Built-in service templates
│   │       ├── postgres.json
│   │       ├── redis.json
│   │       ├── nodejs.json
│   │       ├── nginx.json
│   │       └── ...
│   ├── project/
│   │   ├── ProjectManager.ts    # Save/load .mypaas/project.json
│   │   └── types.ts             # Shared types
│   └── utils/
│       └── portFinder.ts        # Tìm port trống
├── webview/                     # React app cho canvas
│   ├── package.json
│   ├── src/
│   │   ├── App.tsx
│   │   ├── components/
│   │   │   ├── Canvas.tsx           # React Flow canvas
│   │   │   ├── ServiceNode.tsx      # Custom node component
│   │   │   ├── ConnectionEdge.tsx   # Custom edge component
│   │   │   ├── Sidebar.tsx          # Service template list
│   │   │   ├── ConfigPanel.tsx      # Service config editor
│   │   │   ├── LogViewer.tsx        # Realtime log display
│   │   │   ├── TerminalPanel.tsx    # Container shell
│   │   │   └── StatusBar.tsx        # Container stats
│   │   ├── hooks/
│   │   │   ├── useVSCode.ts        # postMessage API
│   │   │   ├── useDocker.ts        # Docker state
│   │   │   └── useCanvas.ts        # Canvas state
│   │   ├── store/
│   │   │   └── canvasStore.ts      # Zustand store
│   │   └── types/
│   │       └── index.ts
│   └── vite.config.ts
└── test/
    └── ...
```

---

## Giao tiếp Webview ↔ Extension

VS Code Webview giao tiếp qua `postMessage`:

```typescript
// Webview → Extension
vscode.postMessage({ type: 'docker:start', serviceId: 'postgres-1' });
vscode.postMessage({ type: 'docker:logs', serviceId: 'api-1' });
vscode.postMessage({ type: 'compose:generate' });
vscode.postMessage({ type: 'compose:up' });
vscode.postMessage({ type: 'project:save', state: canvasState });

// Extension → Webview
panel.webview.postMessage({ type: 'docker:status', data: containerStatuses });
panel.webview.postMessage({ type: 'docker:log', serviceId: 'api-1', line: '...' });
panel.webview.postMessage({ type: 'docker:stats', data: { cpu: 12, ram: 128 } });
panel.webview.postMessage({ type: 'project:loaded', state: savedState });
```

---

## Output: docker-compose.yml tự động sinh

Khi user thiết kế xong trên canvas, extension tự sinh file:

```yaml
# Auto-generated by My PaaS Extension
# Project: cashion-luxury-jewelry

version: "3.8"

services:
  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: mydb
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - app-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - app-network

  api:
    build:
      context: ./api
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgresql://postgres:postgres@postgres:5432/mydb
      REDIS_URL: redis://redis:6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - api
    networks:
      - app-network

volumes:
  postgres-data:
  redis-data:

networks:
  app-network:
    driver: bridge
```

---

## Features theo Phase

### Phase 1 — MVP (2-3 tuần)

- [ ] Extension scaffold + Webview với React Flow
- [ ] Sidebar với 5-8 service templates cơ bản (Postgres, Redis, Node.js, Nginx, MongoDB, MinIO)
- [ ] Kéo thả node lên canvas
- [ ] Click node → edit config (image, ports, env vars)
- [ ] Kéo edge giữa 2 node (tạo depends_on + network)
- [ ] Generate docker-compose.yml từ canvas
- [ ] Nút "Compose Up" / "Compose Down"
- [ ] Save/Load project (.mypaas/project.json)

### Phase 2 — Docker Management (2 tuần)

- [ ] Hiển thị trạng thái container realtime trên mỗi node (running/stopped/error)
- [ ] Stream logs trong panel dưới
- [ ] Terminal attach vào container
- [ ] Container stats (CPU/RAM) hiển thị trên node
- [ ] Start/Stop/Restart từng service riêng lẻ
- [ ] Auto-detect Docker running hay chưa

### Phase 3 — Nâng cao (2-3 tuần)

- [ ] Custom templates: user tạo template riêng
- [ ] Import từ docker-compose.yml có sẵn → render lên canvas
- [ ] Export canvas thành ảnh PNG (documentation)
- [ ] Volume browser: xem nội dung volume
- [ ] Network inspector: xem traffic giữa containers
- [ ] Multi-project support
- [ ] Git integration: auto-commit docker-compose.yml changes

### Phase 4 — Remote & Deploy (tùy chọn)

- [ ] Connect đến Docker Engine remote (SSH / TCP)
- [ ] Deploy lên VPS qua SSH
- [ ] Traefik template với auto SSL
- [ ] Nixpacks integration: build app không cần Dockerfile
- [ ] GitHub repo clone + deploy

---

## Data Model

### CanvasState (lưu trong .mypaas/project.json)

```typescript
interface CanvasState {
  version: string;
  name: string;
  nodes: ServiceNode[];
  edges: ServiceEdge[];
}

interface ServiceNode {
  id: string;
  templateId: string;       // "postgres", "redis", "nodejs-app"
  position: { x: number; y: number };
  data: {
    name: string;           // tên service trong compose
    image?: string;
    build?: { context: string; dockerfile: string };
    ports: string[];        // ["5432:5432"]
    environment: Record<string, string>;
    volumes: string[];
    healthcheck?: HealthcheckConfig;
    labels?: Record<string, string>;
    command?: string;
    restart?: string;       // "always" | "unless-stopped" | "no"
  };
}

interface ServiceEdge {
  id: string;
  source: string;     // node id
  target: string;     // node id
  type: 'depends_on' | 'network';
}
```

---

## Ví dụ Use Case thực tế

### Use Case 1: Setup stack cho dự án web

1. Mở VS Code, nhấn `Ctrl+Shift+P` → "My PaaS: Open Canvas"
2. Kéo **PostgreSQL** từ sidebar → canvas
3. Kéo **Redis** → canvas
4. Kéo **Node.js App** → canvas, trỏ build context tới `./api`
5. Kéo **Nginx** → canvas
6. Nối PostgreSQL → Node.js App (auto thêm `DATABASE_URL`)
7. Nối Redis → Node.js App (auto thêm `REDIS_URL`)
8. Nối Node.js App → Nginx (auto config upstream)
9. Nhấn **"Generate Compose"** → file docker-compose.yml được tạo
10. Nhấn **"Compose Up"** → tất cả container chạy
11. Xem logs realtime ngay trong VS Code

### Use Case 2: Thêm service vào stack đang chạy

1. Mở canvas, thấy stack đang chạy (nodes màu xanh)
2. Kéo **pgAdmin** từ sidebar → canvas
3. Nối pgAdmin → PostgreSQL
4. Extension tự sinh lại docker-compose.yml
5. Nhấn **"Apply Changes"** → chỉ pgAdmin được tạo mới, các container khác không bị restart

---

## Smart Features (tự động hóa)

### Auto Environment Variables

Khi nối PostgreSQL → Node.js App, extension tự thêm:
```
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/mydb
```

Khi nối Redis → Node.js App:
```
REDIS_URL=redis://redis:6379
```

Khi nối MongoDB → Node.js App:
```
MONGODB_URI=mongodb://mongo:27017/mydb
```

### Auto Port Conflict Detection

Nếu 2 service cùng map port 3000, extension cảnh báo và suggest port khác.

### Healthcheck Auto-config

Mỗi database template đi kèm healthcheck mặc định, để `depends_on` hoạt động đúng.

---

## Tech Stack tổng kết

| Component | Technology |
|---|---|
| Extension runtime | TypeScript, VS Code Extension API |
| Webview UI | React, React Flow, Tailwind CSS, Zustand |
| Build tool (webview) | Vite |
| Docker interaction | Dockerode (npm package) |
| Compose management | yaml (npm), child_process (docker compose CLI) |
| Project storage | JSON file (.mypaas/project.json) |
| Template system | JSON files |

---

## So sánh với ý tưởng ban đầu

| Tiêu chí | Railway Clone (IDEA.md) | VS Code Extension (IDEA-v2) |
|---|---|---|
| Độ phức tạp | Rất cao (6+ services) | Thấp (1 extension) |
| Thời gian MVP | 2-3 tháng | 2-3 tuần |
| Cần server riêng | Có | Không |
| Target user | Nhiều team | Developer cá nhân / team nhỏ |
| Giá trị thực tế | Lớn nhưng khó hoàn thành | Nhỏ hơn nhưng dùng được ngay |
| Có thể publish | Phải host platform | Publish lên VS Code Marketplace |
| Monetization | SaaS subscription | Freemium extension |

---

## Bước tiếp theo: Bắt đầu code

```bash
# 1. Scaffold extension
npx --package yo --package generator-code -- yo code

# 2. Chọn: New Extension (TypeScript)
# 3. Setup React + Vite cho webview
# 4. Cài dependencies
npm install dockerode @types/dockerode yaml
cd webview && npm install reactflow zustand @tailwindcss/vite
```

**File đầu tiên cần code:** `src/extension.ts` + `webview/src/App.tsx` với React Flow canvas cơ bản.
