package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/model"
)

// Embedded marketplace templates
var marketplaceTemplates = []model.Template{
	{
		ID:          "wordpress",
		Name:        "WordPress",
		Description: "WordPress CMS with MySQL database",
		Icon:        "globe",
		Services: []model.TemplateService{
			{
				Name:   "wordpress-app",
				Type:   "app",
				GitURL: "file:///data/samples/node-app.git",
				Env:    map[string]string{"WORDPRESS_DB_HOST": "wordpress-db", "WORDPRESS_DB_NAME": "wordpress"},
			},
			{
				Name:  "wordpress-db",
				Type:  "mysql",
				Image: "mysql:8",
				Env:   map[string]string{"MYSQL_DATABASE": "wordpress", "MYSQL_ROOT_PASSWORD": "changeme"},
			},
		},
	},
	{
		ID:          "node-postgres",
		Name:        "Node.js + PostgreSQL",
		Description: "Node.js application with PostgreSQL database",
		Icon:        "server",
		Services: []model.TemplateService{
			{
				Name:   "node-app",
				Type:   "app",
				GitURL: "file:///data/samples/node-app.git",
				Env:    map[string]string{"DATABASE_URL": "postgresql://postgres:changeme@node-db:5432/app"},
			},
			{
				Name:  "node-db",
				Type:  "postgres",
				Image: "postgres:16-alpine",
				Env:   map[string]string{"POSTGRES_DB": "app", "POSTGRES_PASSWORD": "changeme"},
			},
		},
	},
	{
		ID:          "redis-cache",
		Name:        "Redis Cache",
		Description: "Standalone Redis instance for caching",
		Icon:        "database",
		Services: []model.TemplateService{
			{
				Name:  "redis",
				Type:  "redis",
				Image: "redis:7-alpine",
			},
		},
	},
	{
		ID:          "node-redis",
		Name:        "Node.js + Redis",
		Description: "Node.js application with Redis for sessions/caching",
		Icon:        "zap",
		Services: []model.TemplateService{
			{
				Name:   "node-app",
				Type:   "app",
				GitURL: "file:///data/samples/node-app.git",
				Env:    map[string]string{"REDIS_URL": "redis://node-redis:6379"},
			},
			{
				Name:  "node-redis",
				Type:  "redis",
				Image: "redis:7-alpine",
			},
		},
	},
	{
		ID:          "postgres-standalone",
		Name:        "PostgreSQL",
		Description: "Standalone PostgreSQL database",
		Icon:        "database",
		Services: []model.TemplateService{
			{
				Name:  "postgres",
				Type:  "postgres",
				Image: "postgres:16-alpine",
				Env:   map[string]string{"POSTGRES_DB": "mydb", "POSTGRES_PASSWORD": "changeme"},
			},
		},
	},
	{
		ID:          "minio-storage",
		Name:        "MinIO Object Storage",
		Description: "S3-compatible object storage",
		Icon:        "hard-drive",
		Services: []model.TemplateService{
			{
				Name:  "minio",
				Type:  "minio",
				Image: "minio/minio:latest",
				Env:   map[string]string{"MINIO_ROOT_USER": "admin", "MINIO_ROOT_PASSWORD": "changeme123"},
			},
		},
	},
}

func (h *Handler) ListTemplates(c *fiber.Ctx) error {
	return c.JSON(marketplaceTemplates)
}

func (h *Handler) DeployTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	var template *model.Template
	for _, t := range marketplaceTemplates {
		if t.ID == templateID {
			template = &t
			break
		}
	}
	if template == nil {
		return c.Status(404).JSON(fiber.Map{"error": "template not found"})
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil || input.Name == "" {
		input.Name = template.Name
	}

	// Create services from template
	var createdServices []any
	for _, ts := range template.Services {
		if ts.Type == "app" {
			// Create a project for the app service
			p, err := h.Store.CreateProject(model.CreateProjectInput{
				Name:   input.Name + "-" + ts.Name,
				GitURL: ts.GitURL,
			})
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to create project: " + err.Error()})
			}
			// Set env vars
			if len(ts.Env) > 0 {
				var envVars []model.EnvVarInput
				for k, v := range ts.Env {
					envVars = append(envVars, model.EnvVarInput{Key: k, Value: v})
				}
				h.Store.SetEnvVars(p.ID, envVars)
			}
			createdServices = append(createdServices, p)
		} else {
			svc, err := h.Store.CreateService(model.CreateServiceInput{
				Name:  input.Name + "-" + ts.Name,
				Type:  model.ServiceType(ts.Type),
				Image: ts.Image,
			})
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to create service: " + err.Error()})
			}

			// Auto-start the database/service after creation
			go h.autoStartService(svc)

			createdServices = append(createdServices, svc)
		}
	}

	h.Store.AddAuditLog(userID, username, "create", "template", templateID, "deployed template: "+template.Name)
	return c.Status(201).JSON(fiber.Map{
		"message":  "template deployed",
		"template": template.Name,
		"created":  createdServices,
	})
}

// autoStartService pulls the image and starts the container for a backing service.
// Runs in a goroutine so the deploy API responds quickly.
func (h *Handler) autoStartService(svc *model.Service) {
	ctx := context.Background()

	_, defaultEnv, port := model.ServiceDefaults(svc.Type)
	containerName := "mypaas-svc-" + svc.Name

	// Pull image
	if err := h.Docker.PullImage(ctx, svc.Image); err != nil {
		h.Store.UpdateServiceStatus(svc.ID, "error", "")
		return
	}

	labels := map[string]string{
		"mypaas.service": svc.ID,
		"mypaas.type":    string(svc.Type),
	}

	if h.Docker.IsSwarmActive(ctx) {
		envMap := make(map[string]string)
		for k, v := range defaultEnv {
			envMap[k] = v
		}
		serviceID, err := h.Docker.CreateSwarmService(ctx, docker.SwarmServiceOpts{
			Name:     containerName,
			Image:    svc.Image,
			Env:      envMap,
			Labels:   labels,
			Network:  serviceNetwork,
			Replicas: 1,
		})
		if err != nil {
			h.Store.UpdateServiceStatus(svc.ID, "error", "")
			return
		}
		h.Store.UpdateServiceStatus(svc.ID, "running", serviceID)
		return
	}

	// Container mode
	containerID, err := h.Docker.RunContainer(ctx, docker.RunContainerOpts{
		Name:    containerName,
		Image:   svc.Image,
		Env:     defaultEnv,
		Ports:   []string{port},
		Network: serviceNetwork,
		Labels:  labels,
	})
	if err != nil {
		h.Store.UpdateServiceStatus(svc.ID, "error", "")
		return
	}
	h.Store.UpdateServiceStatus(svc.ID, "running", containerID)
}
