# Stage 1: Build Frontend
FROM node:22-alpine AS frontend
WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN chmod -R +x node_modules/.bin && npm run build

# Stage 2: Build Backend
FROM golang:1-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /build

# Copy core module first
COPY core/ /build/core/

# Copy server module
COPY server/go.mod server/go.sum /build/server/
WORKDIR /build/server
RUN go mod download || true

COPY server/ /build/server/
RUN go mod tidy
RUN CGO_ENABLED=1 go build -ldflags="-w -s" -o /mypaas-server .

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates git docker-cli
WORKDIR /app
COPY --from=builder /mypaas-server .
COPY --from=frontend /build/dist ./static/

# Copy sample projects
COPY samples/ ./samples/
RUN chmod +x ./samples/init-samples.sh

RUN mkdir -p /data/builds

EXPOSE 8080
ENV MYPAAS_DB=/data/mypaas.db
ENV MYPAAS_LISTEN=:8080

# Init sample git repos then start server
CMD ["sh", "-c", "./samples/init-samples.sh && ./mypaas-server"]
