package handler

import (
	"github.com/my-paas/server/cache"
	"github.com/my-paas/server/config"
	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/metrics"
	"github.com/my-paas/server/notify"
	"github.com/my-paas/server/store"
	"github.com/my-paas/server/worker"
)

type Handler struct {
	Store    *store.Store
	Docker   *docker.Client
	Queue    *worker.Queue
	Config   *config.Config
	Redis    *cache.Redis
	Metrics  *metrics.Metrics
	Notifier *notify.Notifier
}

func New(s *store.Store, d *docker.Client, q *worker.Queue) *Handler {
	return &Handler{Store: s, Docker: d, Queue: q}
}

func NewEnterprise(s *store.Store, d *docker.Client, q *worker.Queue, cfg *config.Config, redis *cache.Redis, m *metrics.Metrics, n *notify.Notifier) *Handler {
	return &Handler{
		Store:    s,
		Docker:   d,
		Queue:    q,
		Config:   cfg,
		Redis:    redis,
		Metrics:  m,
		Notifier: n,
	}
}
