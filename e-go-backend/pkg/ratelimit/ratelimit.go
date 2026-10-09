// Package ratelimit provides a small fixed-window rate limiter with a Redis
// backend (shared across instances) and an in-process fallback so the app keeps
// working when Redis is disabled in development.
package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter decides whether an action identified by key is allowed. A nil Limiter
// is treated as "always allow" by callers.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// New returns a Redis-backed limiter when a client is available, otherwise an
// in-process limiter. Redis stays optional, matching the rest of the project.
func New(client *redis.Client) Limiter {
	if client != nil {
		return &redisLimiter{client: client}
	}
	return NewMemoryLimiter()
}

type redisLimiter struct {
	client *redis.Client
}

// Allow increments the key and sets the window TTL on first use. Errors are
// returned so the caller can fail open rather than lock users out.
func (l *redisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := l.client.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}
	return count <= int64(limit), nil
}

// MemoryLimiter is a per-process fixed-window limiter. It is only accurate for
// a single instance and is intended for development / Redis-disabled setups.
type MemoryLimiter struct {
	mu      sync.Mutex
	entries map[string]*memoryEntry
	now     func() time.Time
}

type memoryEntry struct {
	count   int
	resetAt time.Time
}

// NewMemoryLimiter builds an empty in-process limiter.
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{entries: make(map[string]*memoryEntry), now: time.Now}
}

// Allow increments the window counter for key, starting a fresh window when the
// previous one has elapsed.
func (l *MemoryLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.entries[key]
	if !ok || now.After(entry.resetAt) {
		if !ok {
			l.sweepLocked(now)
		}
		l.entries[key] = &memoryEntry{count: 1, resetAt: now.Add(window)}
		return true, nil
	}

	entry.count++
	return entry.count <= limit, nil
}

// sweepLocked drops expired entries so the map cannot grow without bound.
func (l *MemoryLimiter) sweepLocked(now time.Time) {
	if len(l.entries) < 1024 {
		return
	}
	for k, e := range l.entries {
		if now.After(e.resetAt) {
			delete(l.entries, k)
		}
	}
}
