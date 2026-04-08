package watcher

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/my-paas/server/model"
	"github.com/my-paas/server/store"
	"github.com/my-paas/server/worker"
)

const defaultPollInterval = 60 * time.Second

type Watcher struct {
	store    *store.Store
	queue    *worker.Queue
	mu       sync.RWMutex
	projects map[string]*projectWatch // projectID → watch state
	cancel   context.CancelFunc
}

type projectWatch struct {
	LastCommit string
}

func New(s *store.Store, q *worker.Queue) *Watcher {
	return &Watcher{
		store:    s,
		queue:    q,
		projects: make(map[string]*projectWatch),
	}
}

// Start begins polling all projects for git changes.
func (w *Watcher) Start(ctx context.Context) {
	ctx, w.cancel = context.WithCancel(ctx)
	go w.loop(ctx)
	log.Println("[watcher] started git polling (interval:", defaultPollInterval, ")")
}

func (w *Watcher) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *Watcher) loop(ctx context.Context) {
	// Initial delay before first poll
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
		return
	}

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.pollAll(ctx)
		}
	}
}

func (w *Watcher) pollAll(ctx context.Context) {
	projects, err := w.store.ListProjects()
	if err != nil {
		log.Printf("[watcher] error listing projects: %v", err)
		return
	}

	for i := range projects {
		p := &projects[i]
		if p.GitURL == "" || !p.AutoDeploy {
			continue
		}
		// Skip projects that are currently deploying
		if p.Status == "deploying" {
			continue
		}

		w.checkProject(ctx, p)
	}
}

func (w *Watcher) checkProject(ctx context.Context, project *model.Project) {
	remoteHash, err := getRemoteHead(ctx, project.GitURL, project.Branch)
	if err != nil {
		// Don't log on every poll failure — could be network glitch
		return
	}

	if remoteHash == "" {
		return
	}

	w.mu.Lock()
	pw, exists := w.projects[project.ID]
	if !exists {
		// First time seeing this project — record hash but don't deploy
		w.projects[project.ID] = &projectWatch{LastCommit: remoteHash}
		w.mu.Unlock()
		log.Printf("[watcher] tracking project=%s branch=%s commit=%s", project.ID, project.Branch, remoteHash[:7])
		return
	}

	if pw.LastCommit == remoteHash {
		w.mu.Unlock()
		return
	}

	// New commit detected!
	oldHash := pw.LastCommit
	pw.LastCommit = remoteHash
	w.mu.Unlock()

	log.Printf("[watcher] new commit detected for project=%s: %s → %s", project.ID, oldHash[:7], remoteHash[:7])

	// Create deployment and queue it
	deploy, err := w.store.CreateDeployment(project.ID, model.TriggerWebhook)
	if err != nil {
		log.Printf("[watcher] error creating deployment: %v", err)
		return
	}

	w.store.UpdateDeploymentCommit(deploy.ID, remoteHash[:7], "")

	err = w.queue.Enqueue(worker.Job{
		Type:         worker.JobDeploy,
		ProjectID:    project.ID,
		DeploymentID: deploy.ID,
	})
	if err != nil {
		log.Printf("[watcher] error enqueueing deploy: %v", err)
	}
}

// getRemoteHead fetches the latest commit hash from a remote git repo without cloning.
func getRemoteHead(ctx context.Context, gitURL, branch string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", gitURL, "refs/heads/"+branch)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return "", nil
	}

	parts := strings.Fields(line)
	if len(parts) < 1 {
		return "", nil
	}

	return parts[0], nil
}
