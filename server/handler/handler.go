package handler

import (
	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/store"
	"github.com/my-paas/server/worker"
)

type Handler struct {
	Store  *store.Store
	Docker *docker.Client
	Queue  *worker.Queue
}

func New(s *store.Store, d *docker.Client, q *worker.Queue) *Handler {
	return &Handler{Store: s, Docker: d, Queue: q}
}
