package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type JobType string

const (
	JobDeploy   JobType = "deploy"
	JobRollback JobType = "rollback"
)

type Job struct {
	Type         JobType
	ProjectID    string
	DeploymentID string
}

type Queue struct {
	jobs    chan Job
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	handler func(ctx context.Context, job Job)
}

func NewQueue(bufSize int, handler func(ctx context.Context, job Job)) *Queue {
	return &Queue{
		jobs:    make(chan Job, bufSize),
		handler: handler,
	}
}

// Start launches worker goroutines to process jobs.
func (q *Queue) Start(ctx context.Context, workers int) {
	ctx, q.cancel = context.WithCancel(ctx)
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go func(workerID int) {
			defer q.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-q.jobs:
					if !ok {
						return
					}
					log.Printf("[worker-%d] processing %s for project=%s deployment=%s",
						workerID, job.Type, job.ProjectID, job.DeploymentID)
					q.handler(ctx, job)
				}
			}
		}(i)
	}
	log.Printf("[queue] started %d workers", workers)
}

func (q *Queue) Enqueue(job Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

func (q *Queue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
	close(q.jobs)
	q.wg.Wait()
}
