# Pipeline triển khai

## Tổng quan

Khi một deployment được trigger (manual, webhook, hoặc auto-detect), My PaaS thực hiện quy trình 5 bước tự động:

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ 1. Clone │───►│ 2.Detect │───►│ 3. Build │───►│ 4.Deploy │───►│5. Health │
│          │    │          │    │          │    │          │    │  Check   │
│ git clone│    │ core     │    │ docker   │    │ swarm/   │    │ TCP/HTTP │
│ --depth 1│    │ .Detect()│    │  build   │    │ container│    │ timeout  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘
    queued        cloning         detecting       building       deploying → healthy/failed
```

Mỗi bước ghi log real-time, client nhận qua **Server-Sent Events (SSE)**.

## Chi tiết từng bước

### Bước 1: Clone

```bash
git clone --depth 1 --branch <branch> <git_url> <build_dir>
```

- **Shallow clone** (`--depth 1`) — chỉ lấy commit mới nhất, tiết kiệm bandwidth
- Trích xuất `commit hash` và `commit message` từ `git log`
- Build directory: `/data/builds/<deployment_id>/`

### Bước 2: Detect (Language Detection)

Core module chạy 7 provider theo thứ tự ưu tiên:

```
PHP → Go → Java → Rust → Python → Node.js → Static File
```

Mỗi provider kiểm tra file đặc trưng:

| Provider | File kiểm tra | Output |
|---|---|---|
| PHP | `composer.json` | PHP 8.2, Apache/FPM |
| Go | `go.mod` | Go 1.26, module-aware build |
| Java | `pom.xml` / `build.gradle` | JDK 21, Maven/Gradle |
| Rust | `Cargo.toml` | Rust stable, cargo build |
| Python | `requirements.txt` / `Pipfile` / `pyproject.toml` | Python 3.12, pip/pipenv |
| Node.js | `package.json` | Node 22 LTS, npm/yarn/bun |
| Static | `index.html` | Nginx, serve static files |

**Output: `BuildPlan`**

```go
type BuildPlan struct {
    Provider   string   // "node", "python", "go", ...
    Language   string   // "Node.js", "Python", "Go", ...
    Version    string   // "22", "3.12", "1.26", ...
    Framework  string   // "Next.js", "Flask", "Fiber", ...
    BaseImage  string   // "node:22-alpine"
    InstallCmd string   // "npm ci"
    BuildCmd   string   // "npm run build"
    StartCmd   string   // "npm start"
    Ports      []int    // [3000]
    EnvVars    []string // ["NODE_ENV=production"]
}
```

### Bước 3: Build (Docker Image)

Từ `BuildPlan`, core module sinh **multi-stage Dockerfile** tối ưu:

```dockerfile
# Ví dụ cho Node.js project
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app .
EXPOSE 3000
ENV NODE_ENV=production
CMD ["npm", "start"]
```

Build command:

```bash
docker build -t mypaas-<project-name>:<commit-hash-8char> .
```

**Cache strategy:**
- Layer caching qua Docker build cache
- `COPY package*.json` trước `COPY .` — chỉ reinstall khi dependencies thay đổi
- `--cache-from` tag `:latest` để reuse layer từ build trước

### Bước 4: Deploy

**Container mode (standalone):**

```
1. docker run (new container)
2. Health check (60s timeout)
3. ✅ → docker stop (old container) → mark healthy
4. ❌ → docker stop (new container) → mark failed
```

**Swarm mode (cluster):**

```
1. docker service create / update (new image tag)
2. Swarm rolling update (start-first strategy)
3. Health check (90s timeout, task state monitoring)
4. ✅ → mark healthy
5. ❌ → mark failed
```

Swarm service config:
- **Replicas**: dựa trên `project.replicas` (default 1)
- **Placement**: `node.role == manager` (default)
- **Network**: `mypaas-network` (overlay, attachable)
- **Labels**: Traefik routing labels tự động gắn

### Bước 5: Health Check

Sau deploy, hệ thống kiểm tra container/service state:

- **Container mode**: Poll `container.inspect()` mỗi 2 giây, đợi state `running`, timeout 60s
- **Swarm mode**: Poll `task.list()` cho service, đợi task state `running`, timeout 90s

Nếu container exit hoặc task fail → deployment đánh dấu `failed`.

## Deployment Status Flow

```
           trigger
             │
             ▼
         ┌─queued─┐
         │        │
         ▼        │ (queue full → retry)
      cloning     │
         │        │
         ▼
     detecting
         │
         ▼
      building
         │
         ▼
     deploying
         │
    ┌────┴────┐
    ▼         ▼
 healthy    failed
```

## Trigger Types

| Type | Source | Chi tiết |
|---|---|---|
| `manual` | Dashboard / CLI / API | User click "Deploy" |
| `webhook` | GitHub push event | `POST /api/webhooks/github` |
| `auto` | Git Watcher | Poll mỗi 60s, so sánh commit hash |
| `env_change` | Environment update | Auto-redeploy khi thay đổi env vars |
| `rollback` | Dashboard / API | Deploy lại image từ deployment trước |

## Rollback

Rollback tìm **deployment healthy gần nhất**, lấy `image_tag`, tạo deployment mới:

```
1. Tìm last healthy deployment (khác deployment hiện tại)
2. Lấy image_tag đã build trước đó
3. Tạo deployment record mới (trigger = "rollback")
4. Skip clone + detect + build → deploy trực tiếp image cũ
5. Health check → healthy/failed
```

Rollback nhanh vì bỏ qua 3 bước đầu — chỉ mất ~10–30 giây.

## Environment Variables

Env vars được inject vào container/service khi deploy:

```
Project envs (DB) + Service link envs (auto-generated) → Container ENV
```

Khi env vars thay đổi qua API, hệ thống tự trigger redeploy (trigger = `env_change`).

**Secret vars** (`is_secret = true`) được mã hoá AES-256-GCM trước khi lưu vào database, chỉ decrypt khi inject vào container.
