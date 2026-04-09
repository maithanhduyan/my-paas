# CLI Tool

## Tổng quan

`mypaas-cli` là công cụ dòng lệnh để quản lý My PaaS từ terminal, phù hợp cho CI/CD pipelines và developer ưa thích terminal workflow.

## Cài đặt

```bash
# Build từ source
cd cli
go build -o mypaas .

# Hoặc cross-compile
GOOS=linux GOARCH=amd64 go build -o mypaas-linux .
GOOS=darwin GOARCH=arm64 go build -o mypaas-mac .
```

## Authentication

### Đăng nhập bằng username/password

```bash
mypaas login --server https://paas.company.com --username admin --password secret
# Logged in to https://paas.company.com as admin (admin)
```

### Đăng nhập bằng API Key

```bash
mypaas login --server https://paas.company.com --api-key mpk_live_xxxxx
# Logged in to https://paas.company.com with API key
```

Credentials lưu tại `~/.mypaas/config.json` (permission `0600`).

### Kiểm tra trạng thái

```bash
mypaas status
# Server:  https://paas.company.com
# Auth:    JWT token

mypaas health
# Status:  ok
# Docker:  connected
# Go:      go1.26.2
```

## Commands

### Projects

```bash
# Liệt kê tất cả projects
mypaas ps
# ID        NAME           STATUS   PROVIDER    BRANCH
# c668b3a9  demo-static    stopped  staticfile  main
# 22025650  test-node      healthy  node        main

# Tạo project mới
mypaas create --name my-app --git-url https://github.com/user/repo.git --branch main

# Deploy
mypaas deploy <project-id>

# Xem logs
mypaas logs <project-id>
mypaas logs <project-id> --tail 50
```

### Environment Variables

```bash
# Liệt kê
mypaas env <project-id>
# KEY       VALUE       SECRET
# NODE_ENV  production
# PORT      3000

# Set nhiều biến cùng lúc
mypaas env <project-id> set PORT=3000 NODE_ENV=production DATABASE_URL=postgres://...

# Xoá biến
mypaas env <project-id> delete OLD_VAR
```

### Services & Domains

```bash
mypaas svc                     # Liệt kê services
mypaas domains <project-id>    # Liệt kê domains
```

### Organizations

```bash
mypaas orgs
# ID        NAME            SLUG            PROJECTS  SERVICES  DEPLOYS
# dc302de3  Acme Corp       acme            50        20        500
```

### API Keys

```bash
# Liệt kê
mypaas keys
# ID        NAME      PREFIX            SCOPES       LAST USED
# 343d3a6b  ci-key    mpk_live_d7ec534  *            2026-04-09T...

# Tạo mới
mypaas keys create --name deploy-key --scopes "read,deploy"

# Xoá
mypaas keys delete <key-id>
```

### Thông tin hệ thống

```bash
mypaas info
# CLI Version: v4.0.0
# Server:      https://paas.company.com
# Go:          go1.26.2
# Docker:      connected

mypaas version
# mypaas-cli v4.0.0
```

## CI/CD Integration

### GitHub Actions

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to My PaaS
        run: |
          mypaas login --server ${{ secrets.PAAS_URL }} --api-key ${{ secrets.PAAS_API_KEY }}
          mypaas deploy ${{ secrets.PROJECT_ID }}
```

### GitLab CI

```yaml
deploy:
  script:
    - mypaas login --server $PAAS_URL --api-key $PAAS_API_KEY
    - mypaas deploy $PROJECT_ID
```

## Config File

CLI lưu config tại `~/.mypaas/config.json`:

```json
{
  "server": "https://paas.company.com",
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_at": ""
}
```

File có permission `0600` (chỉ owner đọc được).

Đăng xuất xoá file: `mypaas logout`
