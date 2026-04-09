package handler

import (
	"github.com/gofiber/fiber/v2"
)

// OpenAPISpec serves the OpenAPI 3.0 JSON specification.
func (h *Handler) OpenAPISpec(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.SendString(openAPIJSON)
}

// SwaggerUI serves a Swagger UI page for interactive API docs.
func (h *Handler) SwaggerUI(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(swaggerHTML)
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>My PaaS API Documentation</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
<style>
body { margin: 0; }
.swagger-ui .topbar { display: none; }
</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
SwaggerUIBundle({
  url: '/api/docs/openapi.json',
  dom_id: '#swagger-ui',
  deepLinking: true,
  presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
  layout: 'BaseLayout'
});
</script>
</body>
</html>`

const openAPIJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "My PaaS Enterprise API",
    "description": "Enterprise-grade PaaS platform API. Supports JWT, API key, and session token authentication.",
    "version": "4.0.0",
    "contact": { "name": "My PaaS Team" },
    "license": { "name": "MIT" }
  },
  "servers": [
    { "url": "/api", "description": "Current server" }
  ],
  "security": [
    { "bearerAuth": [] },
    { "apiKeyAuth": [] }
  ],
  "tags": [
    { "name": "Auth", "description": "Authentication & authorization" },
    { "name": "Projects", "description": "Project management" },
    { "name": "Deployments", "description": "Deploy & rollback" },
    { "name": "Services", "description": "Managed services (PostgreSQL, Redis, etc.)" },
    { "name": "Domains", "description": "Custom domain management" },
    { "name": "Environment", "description": "Environment variables" },
    { "name": "Logs", "description": "Log streaming (SSE)" },
    { "name": "Volumes", "description": "Persistent volumes" },
    { "name": "Stats", "description": "Container & system stats" },
    { "name": "Backups", "description": "Backup & restore" },
    { "name": "Marketplace", "description": "One-click templates" },
    { "name": "Swarm", "description": "Docker Swarm cluster" },
    { "name": "Organizations", "description": "Multi-tenancy & quotas" },
    { "name": "API Keys", "description": "Programmatic access tokens" },
    { "name": "Notifications", "description": "Alert channels & rules" },
    { "name": "Admin", "description": "User & system administration" },
    { "name": "System", "description": "Health, metrics, samples" }
  ],
  "paths": {
    "/health": {
      "get": {
        "tags": ["System"],
        "summary": "Health check",
        "operationId": "healthCheck",
        "security": [],
        "responses": {
          "200": {
            "description": "Server healthy",
            "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string", "example": "ok" }, "docker": { "type": "string", "example": "connected" }, "go": { "type": "string", "example": "go1.26.2" } } } } }
          }
        }
      }
    },
    "/auth/status": {
      "get": {
        "tags": ["Auth"],
        "summary": "Check authentication status",
        "operationId": "authStatus",
        "security": [],
        "responses": {
          "200": { "description": "Auth status", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/AuthStatus" } } } }
        }
      }
    },
    "/auth/setup": {
      "post": {
        "tags": ["Auth"],
        "summary": "Initial admin setup (first user)",
        "operationId": "setup",
        "security": [],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/LoginInput" } } } },
        "responses": {
          "201": { "description": "Admin created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/LoginResponse" } } } },
          "400": { "description": "Already setup" }
        }
      }
    },
    "/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Login with username and password",
        "operationId": "login",
        "security": [],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/LoginInput" } } } },
        "responses": {
          "200": { "description": "Login successful", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/LoginResponse" } } } },
          "401": { "description": "Invalid credentials" }
        }
      }
    },
    "/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Register with invitation token",
        "operationId": "registerWithInvite",
        "security": [],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["token", "username", "password"], "properties": { "token": { "type": "string" }, "username": { "type": "string" }, "password": { "type": "string" } } } } } },
        "responses": {
          "201": { "description": "User registered" },
          "400": { "description": "Invalid or expired token" }
        }
      }
    },
    "/auth/logout": {
      "post": {
        "tags": ["Auth"],
        "summary": "Logout (invalidate session)",
        "operationId": "logout",
        "responses": {
          "200": { "description": "Logged out", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } }
        }
      }
    },
    "/auth/refresh": {
      "post": {
        "tags": ["Auth"],
        "summary": "Refresh access token",
        "operationId": "refreshToken",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["refresh_token"], "properties": { "refresh_token": { "type": "string" } } } } } },
        "responses": {
          "200": { "description": "New access token", "content": { "application/json": { "schema": { "type": "object", "properties": { "access_token": { "type": "string" }, "expires_in": { "type": "integer" } } } } } },
          "401": { "description": "Invalid refresh token" }
        }
      }
    },
    "/projects": {
      "get": {
        "tags": ["Projects"],
        "summary": "List all projects",
        "operationId": "listProjects",
        "responses": {
          "200": { "description": "Project list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Project" } } } } }
        }
      },
      "post": {
        "tags": ["Projects"],
        "summary": "Create a new project",
        "operationId": "createProject",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateProjectInput" } } } },
        "responses": {
          "201": { "description": "Project created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Project" } } } },
          "400": { "description": "Invalid input" }
        }
      }
    },
    "/projects/{id}": {
      "get": {
        "tags": ["Projects"],
        "summary": "Get project by ID",
        "operationId": "getProject",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Project details", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Project" } } } },
          "404": { "description": "Not found" }
        }
      },
      "put": {
        "tags": ["Projects"],
        "summary": "Update project",
        "operationId": "updateProject",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UpdateProjectInput" } } } },
        "responses": {
          "200": { "description": "Project updated", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Project" } } } }
        }
      },
      "delete": {
        "tags": ["Projects"],
        "summary": "Delete project",
        "operationId": "deleteProject",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } }
        }
      }
    },
    "/projects/{id}/deploy": {
      "post": {
        "tags": ["Deployments"],
        "summary": "Trigger deployment",
        "operationId": "triggerDeploy",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "202": { "description": "Deployment queued", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Deployment" } } } }
        }
      }
    },
    "/projects/{id}/deployments": {
      "get": {
        "tags": ["Deployments"],
        "summary": "List deployments for a project",
        "operationId": "listDeployments",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Deployment list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Deployment" } } } } }
        }
      }
    },
    "/deployments/{id}": {
      "get": {
        "tags": ["Deployments"],
        "summary": "Get deployment by ID",
        "operationId": "getDeployment",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Deployment details", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Deployment" } } } }
        }
      }
    },
    "/deployments/{id}/rollback": {
      "post": {
        "tags": ["Deployments"],
        "summary": "Rollback to this deployment",
        "operationId": "rollbackDeployment",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "202": { "description": "Rollback queued", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } }
        }
      }
    },
    "/projects/{id}/restart": {
      "post": {
        "tags": ["Projects"],
        "summary": "Restart project container",
        "operationId": "restartProject",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Restarted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/projects/{id}/stop": {
      "post": {
        "tags": ["Projects"],
        "summary": "Stop project container",
        "operationId": "stopProject",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Stopped", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/projects/{id}/start": {
      "post": {
        "tags": ["Projects"],
        "summary": "Start project container",
        "operationId": "startProject",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Started", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/projects/{id}/env": {
      "get": {
        "tags": ["Environment"],
        "summary": "Get environment variables",
        "operationId": "getEnvVars",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Env vars (secrets masked)", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/EnvVar" } } } } } }
      },
      "put": {
        "tags": ["Environment"],
        "summary": "Set environment variables (bulk)",
        "operationId": "updateEnvVars",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/BulkEnvInput" } } } },
        "responses": { "200": { "description": "Updated", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/projects/{id}/env/{key}": {
      "delete": {
        "tags": ["Environment"],
        "summary": "Delete environment variable",
        "operationId": "deleteEnvVar",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "key", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/projects/{id}/logs": {
      "get": {
        "tags": ["Logs"],
        "summary": "Stream project logs (SSE)",
        "operationId": "streamProjectLogs",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "tail", "in": "query", "schema": { "type": "integer", "default": 100 } }
        ],
        "responses": { "200": { "description": "Server-Sent Events stream", "content": { "text/event-stream": { "schema": { "type": "string" } } } } }
      }
    },
    "/deployments/{id}/logs": {
      "get": {
        "tags": ["Logs"],
        "summary": "Stream deployment logs (SSE)",
        "operationId": "streamDeploymentLogs",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Server-Sent Events stream", "content": { "text/event-stream": { "schema": { "type": "string" } } } } }
      }
    },
    "/services": {
      "get": {
        "tags": ["Services"],
        "summary": "List managed services",
        "operationId": "listServices",
        "responses": { "200": { "description": "Service list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Service" } } } } } }
      },
      "post": {
        "tags": ["Services"],
        "summary": "Create managed service",
        "operationId": "createService",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateServiceInput" } } } },
        "responses": { "201": { "description": "Service created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Service" } } } } }
      }
    },
    "/services/{id}": {
      "delete": {
        "tags": ["Services"],
        "summary": "Delete managed service",
        "operationId": "deleteService",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/services/{id}/start": {
      "post": {
        "tags": ["Services"],
        "summary": "Start service",
        "operationId": "startService",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Started", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/services/{id}/stop": {
      "post": {
        "tags": ["Services"],
        "summary": "Stop service",
        "operationId": "stopService",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Stopped", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/services/{id}/link/{projectId}": {
      "post": {
        "tags": ["Services"],
        "summary": "Link service to project",
        "operationId": "linkService",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "projectId", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": { "content": { "application/json": { "schema": { "type": "object", "properties": { "env_prefix": { "type": "string" } } } } } },
        "responses": { "200": { "description": "Linked", "content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" }, "link": { "type": "object" }, "env_injected": { "type": "integer" } } } } } } }
      },
      "delete": {
        "tags": ["Services"],
        "summary": "Unlink service from project",
        "operationId": "unlinkService",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "projectId", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "Unlinked", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/projects/{id}/domains": {
      "get": {
        "tags": ["Domains"],
        "summary": "List project domains",
        "operationId": "listDomains",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Domain list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Domain" } } } } } }
      },
      "post": {
        "tags": ["Domains"],
        "summary": "Add custom domain",
        "operationId": "addDomain",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateDomainInput" } } } },
        "responses": { "201": { "description": "Domain added", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Domain" } } } } }
      }
    },
    "/domains/{id}": {
      "delete": {
        "tags": ["Domains"],
        "summary": "Delete domain",
        "operationId": "deleteDomain",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/stats": {
      "get": {
        "tags": ["Stats"],
        "summary": "System-wide container stats",
        "operationId": "getSystemStats",
        "responses": { "200": { "description": "Stats", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/ContainerStats" } } } } } }
      }
    },
    "/projects/{id}/stats": {
      "get": {
        "tags": ["Stats"],
        "summary": "Project container stats",
        "operationId": "getProjectStats",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Stats", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ContainerStats" } } } } }
      }
    },
    "/projects/{id}/volumes": {
      "get": {
        "tags": ["Volumes"],
        "summary": "List project volumes",
        "operationId": "listVolumes",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Volume list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Volume" } } } } } }
      },
      "post": {
        "tags": ["Volumes"],
        "summary": "Create persistent volume",
        "operationId": "createVolume",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateVolumeInput" } } } },
        "responses": { "201": { "description": "Volume created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Volume" } } } } }
      }
    },
    "/projects/{id}/volumes/{volumeId}": {
      "delete": {
        "tags": ["Volumes"],
        "summary": "Delete volume",
        "operationId": "deleteVolume",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "volumeId", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/backups": {
      "get": {
        "tags": ["Backups"],
        "summary": "List backups",
        "operationId": "listBackups",
        "responses": { "200": { "description": "Backup list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Backup" } } } } } }
      },
      "post": {
        "tags": ["Backups"],
        "summary": "Create backup (admin)",
        "operationId": "createBackup",
        "responses": { "201": { "description": "Backup created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Backup" } } } } }
      }
    },
    "/backups/{id}/download": {
      "get": {
        "tags": ["Backups"],
        "summary": "Download backup file",
        "operationId": "downloadBackup",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Binary file", "content": { "application/octet-stream": { "schema": { "type": "string", "format": "binary" } } } } }
      }
    },
    "/backups/{id}/restore": {
      "post": {
        "tags": ["Backups"],
        "summary": "Restore from backup (admin)",
        "operationId": "restoreBackup",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Restored", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/backups/{id}": {
      "delete": {
        "tags": ["Backups"],
        "summary": "Delete backup (admin)",
        "operationId": "deleteBackup",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/marketplace": {
      "get": {
        "tags": ["Marketplace"],
        "summary": "List available templates",
        "operationId": "listTemplates",
        "responses": { "200": { "description": "Template list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Template" } } } } } }
      }
    },
    "/marketplace/{id}/deploy": {
      "post": {
        "tags": ["Marketplace"],
        "summary": "Deploy template",
        "operationId": "deployTemplate",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "content": { "application/json": { "schema": { "type": "object", "properties": { "name": { "type": "string" } } } } } },
        "responses": { "201": { "description": "Template deployed" } }
      }
    },
    "/samples": {
      "get": {
        "tags": ["System"],
        "summary": "List sample projects",
        "operationId": "listSamples",
        "responses": { "200": { "description": "Sample list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Sample" } } } } } }
      }
    },
    "/detect": {
      "post": {
        "tags": ["Projects"],
        "summary": "Detect project language/framework",
        "operationId": "detectProject",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["path"], "properties": { "path": { "type": "string" } } } } } },
        "responses": { "200": { "description": "Detection result" } }
      }
    },
    "/swarm/status": {
      "get": {
        "tags": ["Swarm"],
        "summary": "Swarm cluster status",
        "operationId": "swarmStatus",
        "responses": { "200": { "description": "Status", "content": { "application/json": { "schema": { "type": "object", "properties": { "active": { "type": "boolean" }, "nodes": { "type": "array" }, "manager_addr": { "type": "string" } } } } } } }
      }
    },
    "/swarm/services": {
      "get": {
        "tags": ["Swarm"],
        "summary": "List Swarm services",
        "operationId": "swarmServices",
        "responses": { "200": { "description": "Service list" } }
      }
    },
    "/swarm/init": {
      "post": {
        "tags": ["Swarm"],
        "summary": "Initialize Swarm (admin)",
        "operationId": "swarmInit",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["advertise_addr"], "properties": { "advertise_addr": { "type": "string" } } } } } },
        "responses": { "200": { "description": "Swarm initialized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/swarm/token": {
      "get": {
        "tags": ["Swarm"],
        "summary": "Get Swarm join token (admin)",
        "operationId": "swarmToken",
        "responses": { "200": { "description": "Join token", "content": { "application/json": { "schema": { "type": "object", "properties": { "token": { "type": "string" } } } } } } }
      }
    },
    "/users": {
      "get": {
        "tags": ["Admin"],
        "summary": "List users (admin)",
        "operationId": "listUsers",
        "responses": { "200": { "description": "User list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/User" } } } } } }
      }
    },
    "/users/{id}/role": {
      "put": {
        "tags": ["Admin"],
        "summary": "Update user role (admin)",
        "operationId": "updateUserRole",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["role"], "properties": { "role": { "type": "string", "enum": ["admin", "member", "viewer"] } } } } } },
        "responses": { "200": { "description": "Role updated", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/users/{id}": {
      "delete": {
        "tags": ["Admin"],
        "summary": "Delete user (admin)",
        "operationId": "deleteUser",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/invitations": {
      "get": {
        "tags": ["Admin"],
        "summary": "List invitations (admin)",
        "operationId": "listInvitations",
        "responses": { "200": { "description": "Invitation list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Invitation" } } } } } }
      },
      "post": {
        "tags": ["Admin"],
        "summary": "Invite user (admin)",
        "operationId": "inviteUser",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["email", "role"], "properties": { "email": { "type": "string", "format": "email" }, "role": { "type": "string", "enum": ["admin", "member", "viewer"] } } } } } },
        "responses": { "201": { "description": "Invitation created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Invitation" } } } } }
      }
    },
    "/audit": {
      "get": {
        "tags": ["Admin"],
        "summary": "Audit logs (admin)",
        "operationId": "listAuditLogs",
        "responses": { "200": { "description": "Audit log list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/AuditLog" } } } } } }
      }
    },
    "/metrics": {
      "get": {
        "tags": ["System"],
        "summary": "Prometheus metrics",
        "operationId": "getMetrics",
        "security": [],
        "responses": { "200": { "description": "Prometheus text format", "content": { "text/plain": { "schema": { "type": "string" } } } } }
      }
    },
    "/organizations": {
      "get": {
        "tags": ["Organizations"],
        "summary": "List organizations",
        "operationId": "listOrganizations",
        "responses": { "200": { "description": "Organization list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Organization" } } } } } }
      },
      "post": {
        "tags": ["Organizations"],
        "summary": "Create organization",
        "operationId": "createOrganization",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateOrgInput" } } } },
        "responses": { "201": { "description": "Organization created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Organization" } } } } }
      }
    },
    "/organizations/{id}": {
      "get": {
        "tags": ["Organizations"],
        "summary": "Get organization",
        "operationId": "getOrganization",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Organization", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Organization" } } } } }
      },
      "put": {
        "tags": ["Organizations"],
        "summary": "Update organization",
        "operationId": "updateOrganization",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UpdateOrgInput" } } } },
        "responses": { "200": { "description": "Updated", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Organization" } } } } }
      },
      "delete": {
        "tags": ["Organizations"],
        "summary": "Delete organization (admin)",
        "operationId": "deleteOrganization",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/organizations/{id}/members": {
      "get": {
        "tags": ["Organizations"],
        "summary": "List organization members",
        "operationId": "listOrgMembers",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Member list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/OrgMember" } } } } } }
      },
      "post": {
        "tags": ["Organizations"],
        "summary": "Add member to organization",
        "operationId": "addOrgMember",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/AddOrgMemberInput" } } } },
        "responses": { "200": { "description": "Member added", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/organizations/{id}/members/{userId}": {
      "delete": {
        "tags": ["Organizations"],
        "summary": "Remove member from organization",
        "operationId": "removeOrgMember",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "userId", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "Removed", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/organizations/{id}/quotas": {
      "get": {
        "tags": ["Organizations"],
        "summary": "Get organization quota usage",
        "operationId": "getOrgQuotaUsage",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Quota usage", "content": { "application/json": { "schema": { "type": "object", "properties": { "projects": { "$ref": "#/components/schemas/QuotaUsage" }, "services": { "$ref": "#/components/schemas/QuotaUsage" }, "deployments": { "$ref": "#/components/schemas/QuotaUsage" }, "cpu": { "$ref": "#/components/schemas/QuotaUsage" }, "memory": { "$ref": "#/components/schemas/QuotaUsage" } } } } } }
        }
      }
    },
    "/api-keys": {
      "get": {
        "tags": ["API Keys"],
        "summary": "List API keys",
        "operationId": "listAPIKeys",
        "responses": { "200": { "description": "API key list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/APIKey" } } } } } }
      },
      "post": {
        "tags": ["API Keys"],
        "summary": "Create API key",
        "operationId": "createAPIKey",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateAPIKeyInput" } } } },
        "responses": {
          "201": { "description": "API key created (raw key shown only once)", "content": { "application/json": { "schema": { "type": "object", "properties": { "id": { "type": "string" }, "name": { "type": "string" }, "key": { "type": "string", "description": "Full API key — save it now, cannot be retrieved again" }, "key_prefix": { "type": "string" }, "scopes": { "type": "string" }, "created_at": { "type": "string", "format": "date-time" }, "message": { "type": "string" } } } } } }
        }
      }
    },
    "/api-keys/{id}": {
      "delete": {
        "tags": ["API Keys"],
        "summary": "Delete API key",
        "operationId": "deleteAPIKey",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/notifications/channels": {
      "get": {
        "tags": ["Notifications"],
        "summary": "List notification channels",
        "operationId": "listNotificationChannels",
        "parameters": [{ "name": "org_id", "in": "query", "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Channel list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/NotificationChannel" } } } } } }
      },
      "post": {
        "tags": ["Notifications"],
        "summary": "Create notification channel",
        "operationId": "createNotificationChannel",
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateNotificationChannelInput" } } } },
        "responses": { "201": { "description": "Channel created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/NotificationChannel" } } } } }
      }
    },
    "/notifications/channels/{id}": {
      "put": {
        "tags": ["Notifications"],
        "summary": "Enable/disable notification channel",
        "operationId": "updateNotificationChannelStatus",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object", "required": ["enabled"], "properties": { "enabled": { "type": "boolean" } } } } } },
        "responses": { "200": { "description": "Updated", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      },
      "delete": {
        "tags": ["Notifications"],
        "summary": "Delete notification channel",
        "operationId": "deleteNotificationChannel",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/notifications/channels/{channelId}/rules": {
      "get": {
        "tags": ["Notifications"],
        "summary": "List notification rules for channel",
        "operationId": "listNotificationRules",
        "parameters": [{ "name": "channelId", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Rule list", "content": { "application/json": { "schema": { "type": "array", "items": { "$ref": "#/components/schemas/NotificationRule" } } } } } }
      },
      "post": {
        "tags": ["Notifications"],
        "summary": "Create notification rule",
        "operationId": "createNotificationRule",
        "parameters": [{ "name": "channelId", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "$ref": "#/components/schemas/CreateNotificationRuleInput" } } } },
        "responses": { "201": { "description": "Rule created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/NotificationRule" } } } } }
      }
    },
    "/notifications/rules/{ruleId}": {
      "delete": {
        "tags": ["Notifications"],
        "summary": "Delete notification rule",
        "operationId": "deleteNotificationRule",
        "parameters": [{ "name": "ruleId", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "Deleted", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } } } }
      }
    },
    "/webhooks/github": {
      "post": {
        "tags": ["System"],
        "summary": "GitHub webhook receiver",
        "operationId": "githubWebhook",
        "security": [],
        "requestBody": { "required": true, "content": { "application/json": { "schema": { "type": "object" } } } },
        "responses": { "200": { "description": "Webhook processed" } }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "JWT access token or API key (mpk_live_...)"
      },
      "apiKeyAuth": {
        "type": "apiKey",
        "in": "header",
        "name": "Authorization",
        "description": "Bearer mpk_live_xxx API key"
      }
    },
    "schemas": {
      "MessageResponse": {
        "type": "object",
        "properties": { "message": { "type": "string" } }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": { "error": { "type": "string" } }
      },
      "QuotaUsage": {
        "type": "object",
        "properties": { "used": { "type": "number" }, "limit": { "type": "number" } }
      },
      "AuthStatus": {
        "type": "object",
        "properties": {
          "authenticated": { "type": "boolean" },
          "setup_required": { "type": "boolean" },
          "user": { "$ref": "#/components/schemas/User" }
        }
      },
      "LoginInput": {
        "type": "object",
        "required": ["username", "password"],
        "properties": { "username": { "type": "string" }, "password": { "type": "string" } }
      },
      "LoginResponse": {
        "type": "object",
        "properties": {
          "user": { "$ref": "#/components/schemas/User" },
          "token": { "type": "string", "description": "Session token" },
          "expires": { "type": "string", "format": "date-time" },
          "access_token": { "type": "string", "description": "JWT access token" },
          "refresh_token": { "type": "string", "description": "JWT refresh token" },
          "expires_in": { "type": "integer", "description": "Access token TTL in seconds" }
        }
      },
      "User": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "username": { "type": "string" },
          "role": { "type": "string", "enum": ["admin", "member", "viewer"] },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "Project": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "git_url": { "type": "string" },
          "branch": { "type": "string" },
          "provider": { "type": "string" },
          "framework": { "type": "string" },
          "auto_deploy": { "type": "boolean" },
          "status": { "type": "string", "enum": ["active", "ready_to_deploy", "deploying", "healthy", "failed", "stopped"] },
          "cpu_limit": { "type": "number" },
          "mem_limit": { "type": "integer" },
          "replicas": { "type": "integer" },
          "created_by": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" },
          "updated_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateProjectInput": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": { "type": "string" },
          "git_url": { "type": "string" },
          "branch": { "type": "string", "default": "main" }
        }
      },
      "UpdateProjectInput": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "branch": { "type": "string" },
          "auto_deploy": { "type": "boolean" },
          "cpu_limit": { "type": "number" },
          "mem_limit": { "type": "integer" },
          "replicas": { "type": "integer" }
        }
      },
      "Deployment": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "project_id": { "type": "string" },
          "commit_hash": { "type": "string" },
          "commit_msg": { "type": "string" },
          "status": { "type": "string", "enum": ["queued", "cloning", "detecting", "building", "deploying", "healthy", "failed", "rolled_back", "cancelled"] },
          "image_tag": { "type": "string" },
          "trigger": { "type": "string", "enum": ["manual", "webhook", "env_change", "rollback"] },
          "started_at": { "type": "string", "format": "date-time" },
          "finished_at": { "type": "string", "format": "date-time" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "Service": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["postgres", "redis", "mysql", "mongo", "minio"] },
          "image": { "type": "string" },
          "status": { "type": "string", "enum": ["running", "stopped", "error"] },
          "container_id": { "type": "string" },
          "config": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateServiceInput": {
        "type": "object",
        "required": ["name", "type"],
        "properties": {
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["postgres", "redis", "mysql", "mongo", "minio"] },
          "image": { "type": "string" }
        }
      },
      "Domain": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "project_id": { "type": "string" },
          "domain": { "type": "string" },
          "ssl_auto": { "type": "boolean" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateDomainInput": {
        "type": "object",
        "required": ["domain"],
        "properties": {
          "domain": { "type": "string" },
          "ssl_auto": { "type": "boolean", "default": true }
        }
      },
      "EnvVar": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "project_id": { "type": "string" },
          "key": { "type": "string" },
          "value": { "type": "string", "description": "Masked as *** for secrets" },
          "is_secret": { "type": "boolean" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "BulkEnvInput": {
        "type": "object",
        "required": ["vars"],
        "properties": {
          "vars": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["key", "value"],
              "properties": {
                "key": { "type": "string" },
                "value": { "type": "string" },
                "is_secret": { "type": "boolean", "default": false }
              }
            }
          }
        }
      },
      "Volume": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "mount_path": { "type": "string" },
          "project_id": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateVolumeInput": {
        "type": "object",
        "required": ["name", "mount_path"],
        "properties": {
          "name": { "type": "string" },
          "mount_path": { "type": "string" }
        }
      },
      "Backup": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "type": { "type": "string" },
          "service_id": { "type": "string" },
          "filename": { "type": "string" },
          "size": { "type": "integer" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "Template": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "description": { "type": "string" },
          "icon": { "type": "string" },
          "services": { "type": "array", "items": { "type": "object" } }
        }
      },
      "Sample": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "description": { "type": "string" },
          "language": { "type": "string" },
          "icon": { "type": "string" },
          "git_url": { "type": "string" }
        }
      },
      "ContainerStats": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "id": { "type": "string" },
          "cpu_percent": { "type": "number" },
          "mem_usage": { "type": "integer" },
          "mem_limit": { "type": "integer" },
          "mem_percent": { "type": "number" },
          "net_input": { "type": "integer" },
          "net_output": { "type": "integer" },
          "status": { "type": "string" }
        }
      },
      "Organization": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "slug": { "type": "string" },
          "max_projects": { "type": "integer" },
          "max_services": { "type": "integer" },
          "max_cpu": { "type": "number" },
          "max_memory": { "type": "integer" },
          "max_deployments": { "type": "integer" },
          "created_at": { "type": "string", "format": "date-time" },
          "updated_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateOrgInput": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": { "type": "string" },
          "slug": { "type": "string" },
          "max_projects": { "type": "integer", "default": 0 },
          "max_services": { "type": "integer", "default": 0 },
          "max_cpu": { "type": "number", "default": 0 },
          "max_memory": { "type": "integer", "default": 0 },
          "max_deployments": { "type": "integer", "default": 0 }
        }
      },
      "UpdateOrgInput": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "max_projects": { "type": "integer" },
          "max_services": { "type": "integer" },
          "max_cpu": { "type": "number" },
          "max_memory": { "type": "integer" },
          "max_deployments": { "type": "integer" }
        }
      },
      "OrgMember": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "org_id": { "type": "string" },
          "user_id": { "type": "string" },
          "role": { "type": "string", "enum": ["owner", "admin", "member", "viewer"] },
          "username": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "AddOrgMemberInput": {
        "type": "object",
        "required": ["user_id", "role"],
        "properties": {
          "user_id": { "type": "string" },
          "role": { "type": "string", "enum": ["owner", "admin", "member", "viewer"] }
        }
      },
      "APIKey": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "user_id": { "type": "string" },
          "name": { "type": "string" },
          "key_prefix": { "type": "string" },
          "scopes": { "type": "string" },
          "last_used": { "type": "string", "format": "date-time" },
          "expires_at": { "type": "string", "format": "date-time" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateAPIKeyInput": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": { "type": "string" },
          "scopes": { "type": "string", "default": "*", "description": "Comma-separated: read,deploy,admin or * for all" }
        }
      },
      "NotificationChannel": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "org_id": { "type": "string" },
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["webhook", "slack", "email"] },
          "config": { "type": "string", "description": "JSON config string" },
          "enabled": { "type": "boolean" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateNotificationChannelInput": {
        "type": "object",
        "required": ["name", "type", "config"],
        "properties": {
          "org_id": { "type": "string" },
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["webhook", "slack", "email"] },
          "config": { "type": "string" },
          "enabled": { "type": "boolean", "default": true }
        }
      },
      "NotificationRule": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "channel_id": { "type": "string" },
          "event": { "type": "string", "enum": ["deploy.started", "deploy.succeeded", "deploy.failed", "health.down", "health.recovered", "quota.warning", "backup.completed"] },
          "project_id": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateNotificationRuleInput": {
        "type": "object",
        "required": ["event"],
        "properties": {
          "event": { "type": "string", "enum": ["deploy.started", "deploy.succeeded", "deploy.failed", "health.down", "health.recovered", "quota.warning", "backup.completed"] },
          "project_id": { "type": "string" }
        }
      },
      "AuditLog": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "user_id": { "type": "string" },
          "username": { "type": "string" },
          "action": { "type": "string" },
          "resource": { "type": "string" },
          "resource_id": { "type": "string" },
          "details": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" }
        }
      },
      "Invitation": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "email": { "type": "string" },
          "role": { "type": "string" },
          "token": { "type": "string" },
          "used": { "type": "boolean" },
          "created_by": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" },
          "expires_at": { "type": "string", "format": "date-time" }
        }
      }
    }
  }
}`
