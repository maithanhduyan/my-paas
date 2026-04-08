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

	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/handler"
	"github.com/my-paas/server/middleware"
	"github.com/my-paas/server/store"
	"github.com/my-paas/server/watcher"
	"github.com/my-paas/server/worker"
)

func main() {
	// Config
	dbPath := envOr("MYPAAS_DB", "/data/mypaas.db")
	listen := envOr("MYPAAS_LISTEN", ":8080")

	// Ensure data directory exists
	os.MkdirAll("/data/builds", 0o755)

	// Init store
	db, err := store.New(dbPath)
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

	// Init deploy worker
	deployWorker := &worker.DeployWorker{
		Store:  db,
		Docker: dockerClient,
	}

	// Init job queue (buffer=100, 2 concurrent workers)
	queue := worker.NewQueue(100, deployWorker.Handle)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx, 2)

	// Init handlers
	h := handler.New(db, dockerClient, queue)

	// Start git watcher
	gitWatcher := watcher.New(db, queue)
	gitWatcher.Start(ctx)

	// Init Fiber
	app := fiber.New(fiber.Config{
		AppName:      "My PaaS Server",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} ${method} ${path} ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	// Routes
	api := app.Group("/api")

	// --- Public routes (no auth required) ---
	api.Get("/health", h.HealthCheck)
	api.Get("/auth/status", h.AuthStatus)
	api.Post("/auth/setup", h.Setup)
	api.Post("/auth/login", h.Login)
	api.Post("/auth/register", h.RegisterWithInvite)
	api.Post("/webhooks/github", h.GithubWebhook)

	// --- Protected routes ---
	protected := api.Group("", middleware.AuthRequired(db))

	protected.Post("/auth/logout", h.Logout)

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
	admin.Post("/swarm/init", h.SwarmInit)
	admin.Get("/swarm/token", h.SwarmToken)

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

	log.Printf("My PaaS server starting on %s", listen)
	if err := app.Listen(listen); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
