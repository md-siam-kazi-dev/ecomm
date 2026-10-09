package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiter_BlocksAndResets(t *testing.T) {
	limiter := NewMemoryLimiter()
	now := time.Unix(0, 0)
	limiter.now = func() time.Time { return now }
	ctx := context.Background()

	for i := 1; i <= 2; i++ {
		allowed, err := limiter.Allow(ctx, "k", 2, time.Minute)
		if err != nil || !allowed {
			t.Fatalf("call %d should be allowed (err=%v, allowed=%v)", i, err, allowed)
		}
	}

	if allowed, _ := limiter.Allow(ctx, "k", 2, time.Minute); allowed {
		t.Fatal("third call within the window should be blocked")
	}

	// After the window elapses the counter resets.
	now = now.Add(time.Minute + time.Second)
	if allowed, _ := limiter.Allow(ctx, "k", 2, time.Minute); !allowed {
		t.Fatal("call after the window should be allowed")
	}
}

func TestMemoryLimiter_KeysAreIsolated(t *testing.T) {
	limiter := NewMemoryLimiter()
	ctx := context.Background()

	if allowed, _ := limiter.Allow(ctx, "a", 1, time.Minute); !allowed {
		t.Fatal("first key should be allowed")
	}
	if allowed, _ := limiter.Allow(ctx, "b", 1, time.Minute); !allowed {
		t.Fatal("second key should be allowed independently")
	}
}

func TestNew_NilRedisUsesMemory(t *testing.T) {
	if _, ok := New(nil).(*MemoryLimiter); !ok {
		t.Fatal("expected an in-process limiter when Redis is nil")
	}
}
