package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/my-paas/server/cache"
	"github.com/my-paas/server/config"
	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/handler"
	"github.com/my-paas/server/metrics"
	"github.com/my-paas/server/middleware"
	"github.com/my-paas/server/notify"
	"github.com/my-paas/server/store"
	"github.com/my-paas/server/watcher"
	"github.com/my-paas/server/worker"
)

func main() {
	// Load config
	cfg := config.Load()

	// Ensure data directories exist
	os.MkdirAll(cfg.BuildsDir, 0o755)
	os.MkdirAll(cfg.DataDir, 0o755)

	// Init store (SQLite or PostgreSQL)
	var db *store.Store
	var err error
	switch cfg.DBDriver {
	case "postgres":
		log.Println("Using PostgreSQL database")
		db, err = store.NewPostgres(cfg.DBURL)
	default:
		log.Println("Using SQLite database")
		db, err = store.New(cfg.DBPath)
	}
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	// Init Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to Docker: %v", err)
	}
	defer dockerClient.Close()

	// Verify Docker connection
	if err := dockerClient.Ping(context.Background()); err != nil {
		log.Printf("WARNING: Docker is not reachable: %v", err)
		log.Printf("Make sure Docker socket is mounted at /var/run/docker.sock")
	} else {
		log.Println("Docker connection OK")
	}

	// Init Redis (optional)
	var redis *cache.Redis
	if cfg.RedisEnabled {
		redis, err = cache.NewRedis(cfg.RedisURL)
		if err != nil {
			log.Printf("WARNING: Redis not available: %v (falling back to in-memory)", err)
			redis = nil
		} else {
			log.Println("Redis connection OK")
			defer redis.Close()
		}
	}

	// Init metrics
	m := metrics.New()

	// Init deploy worker
	deployWorker := &worker.DeployWorker{
		Store:  db,
		Docker: dockerClient,
	}

	// Init job queue
	queue := worker.NewQueue(cfg.QueueSize, deployWorker.Handle)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx, cfg.WorkerCount)

	// Init notifier
	notifier := notify.New(db)

	// Init handlers
	h := handler.NewEnterprise(db, dockerClient, queue, cfg, redis, m, notifier)

	// Start git watcher
	gitWatcher := watcher.New(db, queue)
	gitWatcher.Start(ctx)

	// Init Fiber
	app := fiber.New(fiber.Config{
		AppName:      "My PaaS Enterprise Server",
		ErrorHandler: customErrorHandler,
		BodyLimit:    50 * 1024 * 1024, // 50MB for large uploads
	})

	// Middleware stack
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.SecurityHeaders())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} ${method} ${path} ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization,X-Request-ID",
	}))
	app.Use(m.MetricsMiddleware())
	app.Use(middleware.RateLimiter(cfg, redis))

	// Routes
	api := app.Group("/api")

	// --- Public routes (no auth required) ---
	api.Get("/health", h.HealthCheck)
	api.Get("/auth/status", h.AuthStatus)
	api.Post("/auth/setup", h.Setup)
	api.Post("/auth/login", h.Login)
	api.Post("/auth/register", h.RegisterWithInvite)
	api.Post("/webhooks/github", h.GithubWebhook)

	// Prometheus metrics endpoint
	api.Get("/metrics", m.Handler())

	// OpenAPI / Swagger documentation
	api.Get("/docs/openapi.json", h.OpenAPISpec)
	api.Get("/docs", h.SwaggerUI)

	// --- Protected routes ---
	protected := api.Group("", middleware.AuthRequired(db, cfg))

	protected.Post("/auth/logout", h.Logout)
	protected.Post("/auth/refresh", h.RefreshToken)

	// Projects
	protected.Get("/projects", h.ListProjects)
	protected.Post("/projects", h.CreateProject)
	protected.Get("/projects/:id", h.GetProject)
	protected.Put("/projects/:id", h.UpdateProject)
	protected.Delete("/projects/:id", h.DeleteProject)

	// Detection
	protected.Post("/detect", h.DetectProject)

	// Deployments
	protected.Post("/projects/:id/deploy", h.TriggerDeploy)
	protected.Get("/projects/:id/deployments", h.ListDeployments)
	protected.Get("/deployments/:id", h.GetDeployment)
	protected.Post("/deployments/:id/rollback", h.RollbackDeployment)

	// Project container actions
	protected.Post("/projects/:id/restart", h.RestartProject)
	protected.Post("/projects/:id/stop", h.StopProject)
	protected.Post("/projects/:id/start", h.StartProject)

	// Environment
	protected.Get("/projects/:id/env", h.GetEnvVars)
	protected.Put("/projects/:id/env", h.UpdateEnvVars)
	protected.Delete("/projects/:id/env/:key", h.DeleteEnvVar)

	// Logs (SSE)
	protected.Get("/projects/:id/logs", h.StreamProjectLogs)
	protected.Get("/deployments/:id/logs", h.StreamDeploymentLogs)

	// Services
	protected.Get("/services", h.ListServices)
	protected.Post("/services", h.CreateService)
	protected.Delete("/services/:id", h.DeleteService)
	protected.Post("/services/:id/start", h.StartService)
	protected.Post("/services/:id/stop", h.StopService)
	protected.Post("/services/:id/link/:projectId", h.LinkServiceToProject)
	protected.Delete("/services/:id/link/:projectId", h.UnlinkServiceFromProject)

	// Domains
	protected.Get("/projects/:id/domains", h.ListDomains)
	protected.Post("/projects/:id/domains", h.AddDomain)
	protected.Delete("/domains/:id", h.DeleteDomain)

	// Stats
	protected.Get("/stats", h.GetSystemStats)
	protected.Get("/projects/:id/stats", h.GetProjectStats)

	// Volumes
	protected.Get("/projects/:id/volumes", h.ListVolumes)
	protected.Post("/projects/:id/volumes", h.CreateVolume)
	protected.Delete("/projects/:id/volumes/:volumeId", h.DeleteVolume)

	// Backups (admin only)
	protected.Get("/backups", h.ListBackups)
	protected.Post("/backups", middleware.RoleRequired("admin"), h.CreateBackup)
	protected.Get("/backups/:id/download", h.DownloadBackup)
	protected.Post("/backups/:id/restore", middleware.RoleRequired("admin"), h.RestoreBackup)
	protected.Delete("/backups/:id", middleware.RoleRequired("admin"), h.DeleteBackup)

	// Marketplace
	protected.Get("/marketplace", h.ListTemplates)
	protected.Post("/marketplace/:id/deploy", h.DeployTemplate)

	// Samples
	protected.Get("/samples", h.ListSamples)

	// --- Admin routes ---
	admin := protected.Group("", middleware.RoleRequired("admin"))
	admin.Get("/users", h.ListUsers)
	admin.Put("/users/:id/role", h.UpdateUserRole)
	admin.Delete("/users/:id", h.DeleteUserAccount)
	admin.Post("/invitations", h.InviteUser)
	admin.Get("/invitations", h.ListInvitations)
	admin.Get("/audit", h.ListAuditLogs)

	// --- Swarm routes ---
	protected.Get("/swarm/status", h.SwarmStatus)
	protected.Get("/swarm/services", h.SwarmServices)
	admin.Post("/swarm/init", h.SwarmInit)
	admin.Get("/swarm/token", h.SwarmToken)

	// --- Enterprise routes ---

	// Registry
	protected.Get("/registry/status", h.GetRegistryStatus)
	admin.Post("/registry/start", h.StartRegistry)
	admin.Post("/registry/stop", h.StopRegistry)
	protected.Get("/registry/images", h.ListRegistryImages)
	admin.Delete("/registry/images/:name", h.DeleteRegistryImage)
	protected.Post("/projects/:id/push", h.PushToRegistry)

	// Organizations
	protected.Get("/organizations", h.ListOrganizations)
	protected.Post("/organizations", h.CreateOrganization)
	protected.Get("/organizations/:id", h.GetOrganization)
	protected.Put("/organizations/:id", h.UpdateOrganization)
	protected.Delete("/organizations/:id", middleware.RoleRequired("admin"), h.DeleteOrganization)
	protected.Get("/organizations/:id/members", h.ListOrgMembers)
	protected.Post("/organizations/:id/members", h.AddOrgMember)
	protected.Delete("/organizations/:id/members/:userId", h.RemoveOrgMember)
	protected.Get("/organizations/:id/quotas", h.GetOrgQuotaUsage)

	// API Keys
	protected.Get("/api-keys", h.ListAPIKeys)
	protected.Post("/api-keys", h.CreateAPIKey)
	protected.Delete("/api-keys/:id", h.DeleteAPIKey)

	// Notification Channels
	protected.Get("/notifications/channels", h.ListNotificationChannels)
	protected.Post("/notifications/channels", h.CreateNotificationChannel)
	protected.Put("/notifications/channels/:id", h.UpdateNotificationChannelStatus)
	protected.Delete("/notifications/channels/:id", h.DeleteNotificationChannel)
	protected.Get("/notifications/channels/:channelId/rules", h.ListNotificationRules)
	protected.Post("/notifications/channels/:channelId/rules", h.CreateNotificationRule)
	protected.Delete("/notifications/rules/:ruleId", h.DeleteNotificationRule)

	// Serve static frontend (production)
	app.Static("/", "./static")
	// SPA fallback
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		gitWatcher.Stop()
		queue.Stop()
		app.Shutdown()
	}()

	log.Printf("My PaaS Enterprise server starting on %s (db: %s)", cfg.Listen, cfg.DBDriver)
	if err := app.Listen(cfg.Listen); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
