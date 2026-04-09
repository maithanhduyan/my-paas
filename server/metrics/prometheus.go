package metrics

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Metrics collects and exposes Prometheus-compatible metrics.
type Metrics struct {
	mu sync.RWMutex

	// Counters
	deploymentsTotal    map[string]*int64 // status -> count
	apiRequestsTotal    map[string]*int64 // method:path:status -> count
	authFailuresTotal   int64
	buildSecondsTotal   float64

	// Gauges
	activeProjects  int64
	activeServices  int64
	queueDepth      int64
	workerBusy      int64

	// Histograms (simplified - track sum and count)
	apiLatencySum   map[string]*float64 // method:path -> sum
	apiLatencyCount map[string]*int64
	deployDuration  map[string]*float64 // project -> sum
	deployCount     map[string]*int64

	startTime time.Time
}

func New() *Metrics {
	return &Metrics{
		deploymentsTotal:    make(map[string]*int64),
		apiRequestsTotal:    make(map[string]*int64),
		apiLatencySum:       make(map[string]*float64),
		apiLatencyCount:     make(map[string]*int64),
		deployDuration:      make(map[string]*float64),
		deployCount:         make(map[string]*int64),
		startTime:           time.Now(),
	}
}

func (m *Metrics) IncDeployment(status string) {
	m.mu.Lock()
	if m.deploymentsTotal[status] == nil {
		v := int64(0)
		m.deploymentsTotal[status] = &v
	}
	m.mu.Unlock()
	m.mu.RLock()
	atomic.AddInt64(m.deploymentsTotal[status], 1)
	m.mu.RUnlock()
}

func (m *Metrics) RecordDeployDuration(project string, seconds float64) {
	m.mu.Lock()
	if m.deployDuration[project] == nil {
		v := float64(0)
		m.deployDuration[project] = &v
		c := int64(0)
		m.deployCount[project] = &c
	}
	*m.deployDuration[project] += seconds
	atomic.AddInt64(m.deployCount[project], 1)
	m.mu.Unlock()
}

func (m *Metrics) SetActiveProjects(n int64)  { atomic.StoreInt64(&m.activeProjects, n) }
func (m *Metrics) SetActiveServices(n int64)   { atomic.StoreInt64(&m.activeServices, n) }
func (m *Metrics) SetQueueDepth(n int64)       { atomic.StoreInt64(&m.queueDepth, n) }
func (m *Metrics) SetWorkerBusy(n int64)       { atomic.StoreInt64(&m.workerBusy, n) }
func (m *Metrics) IncAuthFailure()             { atomic.AddInt64(&m.authFailuresTotal, 1) }

func (m *Metrics) RecordAPIRequest(method, path string, status int, latency time.Duration) {
	key := fmt.Sprintf("%s:%s:%d", method, normalizePath(path), status)
	latencyKey := fmt.Sprintf("%s:%s", method, normalizePath(path))
	seconds := latency.Seconds()

	m.mu.Lock()
	if m.apiRequestsTotal[key] == nil {
		v := int64(0)
		m.apiRequestsTotal[key] = &v
	}
	if m.apiLatencySum[latencyKey] == nil {
		v := float64(0)
		m.apiLatencySum[latencyKey] = &v
		c := int64(0)
		m.apiLatencyCount[latencyKey] = &c
	}
	*m.apiLatencySum[latencyKey] += seconds
	atomic.AddInt64(m.apiLatencyCount[latencyKey], 1)
	m.mu.Unlock()

	m.mu.RLock()
	atomic.AddInt64(m.apiRequestsTotal[key], 1)
	m.mu.RUnlock()
}

// Handler returns a Fiber handler that serves Prometheus text format metrics.
func (m *Metrics) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		var b strings.Builder

		// Process info
		b.WriteString("# HELP mypaas_info My PaaS server information\n")
		b.WriteString("# TYPE mypaas_info gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_info{version=\"4.0.0\",go_version=\"%s\"} 1\n", runtime.Version()))

		// Uptime
		b.WriteString("# HELP mypaas_uptime_seconds Server uptime in seconds\n")
		b.WriteString("# TYPE mypaas_uptime_seconds gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_uptime_seconds %f\n", time.Since(m.startTime).Seconds()))

		// Deployments
		b.WriteString("# HELP mypaas_deployments_total Total deployments by status\n")
		b.WriteString("# TYPE mypaas_deployments_total counter\n")
		m.mu.RLock()
		for status, count := range m.deploymentsTotal {
			b.WriteString(fmt.Sprintf("mypaas_deployments_total{status=\"%s\"} %d\n", status, atomic.LoadInt64(count)))
		}
		m.mu.RUnlock()

		// Active gauges
		b.WriteString("# HELP mypaas_active_projects Total active projects\n")
		b.WriteString("# TYPE mypaas_active_projects gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_active_projects %d\n", atomic.LoadInt64(&m.activeProjects)))

		b.WriteString("# HELP mypaas_active_services Total active services\n")
		b.WriteString("# TYPE mypaas_active_services gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_active_services %d\n", atomic.LoadInt64(&m.activeServices)))

		b.WriteString("# HELP mypaas_queue_depth Current job queue depth\n")
		b.WriteString("# TYPE mypaas_queue_depth gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_queue_depth %d\n", atomic.LoadInt64(&m.queueDepth)))

		b.WriteString("# HELP mypaas_worker_busy Current busy workers\n")
		b.WriteString("# TYPE mypaas_worker_busy gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_worker_busy %d\n", atomic.LoadInt64(&m.workerBusy)))

		// Auth failures
		b.WriteString("# HELP mypaas_auth_failures_total Total authentication failures\n")
		b.WriteString("# TYPE mypaas_auth_failures_total counter\n")
		b.WriteString(fmt.Sprintf("mypaas_auth_failures_total %d\n", atomic.LoadInt64(&m.authFailuresTotal)))

		// API requests
		b.WriteString("# HELP mypaas_api_requests_total Total API requests\n")
		b.WriteString("# TYPE mypaas_api_requests_total counter\n")
		m.mu.RLock()
		for key, count := range m.apiRequestsTotal {
			parts := strings.SplitN(key, ":", 3)
			if len(parts) == 3 {
				b.WriteString(fmt.Sprintf("mypaas_api_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
					parts[0], parts[1], parts[2], atomic.LoadInt64(count)))
			}
		}
		m.mu.RUnlock()

		// API latency
		b.WriteString("# HELP mypaas_api_request_duration_seconds API request duration\n")
		b.WriteString("# TYPE mypaas_api_request_duration_seconds summary\n")
		m.mu.RLock()
		for key, sum := range m.apiLatencySum {
			parts := strings.SplitN(key, ":", 2)
			if len(parts) == 2 {
				count := atomic.LoadInt64(m.apiLatencyCount[key])
				b.WriteString(fmt.Sprintf("mypaas_api_request_duration_seconds_sum{method=\"%s\",path=\"%s\"} %f\n", parts[0], parts[1], *sum))
				b.WriteString(fmt.Sprintf("mypaas_api_request_duration_seconds_count{method=\"%s\",path=\"%s\"} %d\n", parts[0], parts[1], count))
			}
		}
		m.mu.RUnlock()

		// Deploy durations
		b.WriteString("# HELP mypaas_deploy_duration_seconds Deploy duration by project\n")
		b.WriteString("# TYPE mypaas_deploy_duration_seconds summary\n")
		m.mu.RLock()
		for project, sum := range m.deployDuration {
			count := atomic.LoadInt64(m.deployCount[project])
			b.WriteString(fmt.Sprintf("mypaas_deploy_duration_seconds_sum{project=\"%s\"} %f\n", project, *sum))
			b.WriteString(fmt.Sprintf("mypaas_deploy_duration_seconds_count{project=\"%s\"} %d\n", project, count))
		}
		m.mu.RUnlock()

		// Go runtime metrics
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		b.WriteString("# HELP mypaas_go_goroutines Current number of goroutines\n")
		b.WriteString("# TYPE mypaas_go_goroutines gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_go_goroutines %d\n", runtime.NumGoroutine()))

		b.WriteString("# HELP mypaas_go_memory_alloc_bytes Current memory allocation in bytes\n")
		b.WriteString("# TYPE mypaas_go_memory_alloc_bytes gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_go_memory_alloc_bytes %d\n", memStats.Alloc))

		b.WriteString("# HELP mypaas_go_memory_sys_bytes Total memory obtained from OS\n")
		b.WriteString("# TYPE mypaas_go_memory_sys_bytes gauge\n")
		b.WriteString(fmt.Sprintf("mypaas_go_memory_sys_bytes %d\n", memStats.Sys))

		return c.SendString(b.String())
	}
}

// MetricsMiddleware records API request metrics.
func (m *Metrics) MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		m.RecordAPIRequest(c.Method(), c.Path(), c.Response().StatusCode(), latency)
		return err
	}
}

// normalizePath replaces ID segments with :id for aggregation.
func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if len(p) == 8 && !strings.Contains(p, ".") {
			parts[i] = ":id"
		}
	}
	return strings.Join(parts, "/")
}
