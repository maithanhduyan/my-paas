package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(url string) (*Redis, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	opts.PoolSize = 20
	opts.MinIdleConns = 5

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

// --- Key-Value Cache ---

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// --- Job Queue (reliable queue using BRPOPLPUSH pattern) ---

const (
	QueuePending    = "mypaas:queue:pending"
	QueueProcessing = "mypaas:queue:processing"
)

type QueueJob struct {
	ID           string `json:"id"`
	Type         string `json:"type"` // "deploy" | "rollback"
	ProjectID    string `json:"project_id"`
	DeploymentID string `json:"deployment_id"`
	CreatedAt    int64  `json:"created_at"`
}

func (r *Redis) EnqueueJob(ctx context.Context, job QueueJob) error {
	job.CreatedAt = time.Now().UnixMilli()
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return r.client.LPush(ctx, QueuePending, data).Err()
}

func (r *Redis) DequeueJob(ctx context.Context, timeout time.Duration) (*QueueJob, error) {
	result, err := r.client.BRPopLPush(ctx, QueuePending, QueueProcessing, timeout).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var job QueueJob
	if err := json.Unmarshal(result, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *Redis) AckJob(ctx context.Context, job QueueJob) error {
	data, _ := json.Marshal(job)
	return r.client.LRem(ctx, QueueProcessing, 1, data).Err()
}

func (r *Redis) QueueDepth(ctx context.Context) (int64, error) {
	return r.client.LLen(ctx, QueuePending).Result()
}

// --- Pub/Sub for real-time events ---

func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

func (r *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// Log channel helpers
func DeployLogChannel(deploymentID string) string {
	return fmt.Sprintf("mypaas:logs:deploy:%s", deploymentID)
}

func ProjectLogChannel(projectID string) string {
	return fmt.Sprintf("mypaas:logs:project:%s", projectID)
}

// --- Distributed Lock ---

func (r *Redis) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("mypaas:lock:%s", key)
	ok, err := r.client.SetNX(ctx, lockKey, "1", ttl).Result()
	return ok, err
}

func (r *Redis) ReleaseLock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("mypaas:lock:%s", key)
	return r.client.Del(ctx, lockKey).Err()
}

// --- Rate Limiting (sliding window) ---

func (r *Redis) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()
	rateLimitKey := fmt.Sprintf("mypaas:ratelimit:%s", key)

	pipe := r.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, rateLimitKey, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, rateLimitKey)
	pipe.ZAdd(ctx, rateLimitKey, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
	pipe.Expire(ctx, rateLimitKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := int(countCmd.Val())
	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	if count >= limit {
		return false, remaining, nil
	}
	return true, remaining, nil
}

// --- Health ---

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// --- Metrics helpers ---

func (r *Redis) IncrCounter(ctx context.Context, key string) error {
	return r.client.Incr(ctx, key).Err()
}

func (r *Redis) GetCounter(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// LogEvent publishes a notification event for async processing
func (r *Redis) LogEvent(ctx context.Context, event string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"event":   event,
		"payload": payload,
		"time":    time.Now().Unix(),
	})
	if err != nil {
		log.Printf("[redis] failed to marshal event: %v", err)
		return
	}
	r.client.Publish(ctx, "mypaas:events", data)
}
