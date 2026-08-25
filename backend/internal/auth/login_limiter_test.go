package auth

import (
	"testing"
	"time"
)

func TestLoginLimiterBlocksAndExpires(t *testing.T) {
	limiter := NewLoginLimiter(3, time.Minute)
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	const key = "192.0.2.1"

	for i := 0; i < 3; i++ {
		if !limiter.Allow(key, now) {
			t.Fatalf("Allow() blocked before failure %d", i+1)
		}
		limiter.Failure(key, now)
	}
	if limiter.Allow(key, now.Add(59*time.Second)) {
		t.Fatal("Allow() accepted a blocked key")
	}
	if !limiter.Allow(key, now.Add(time.Minute)) {
		t.Fatal("Allow() did not release a key after the block duration")
	}
}

func TestLoginLimiterSuccessClearsFailures(t *testing.T) {
	limiter := NewLoginLimiter(2, time.Minute)
	now := time.Now()
	const key = "192.0.2.2"

	limiter.Failure(key, now)
	limiter.Success(key)
	limiter.Failure(key, now)
	if !limiter.Allow(key, now) {
		t.Fatal("Allow() blocked after Success() reset the failure count")
	}
}

func TestLoginLimiterAllowsEmptyKey(t *testing.T) {
	limiter := NewLoginLimiter(1, time.Minute)
	limiter.Failure("", time.Now())
	if !limiter.Allow("", time.Now()) {
		t.Fatal("Allow() should fail open when no client key is available")
	}
}
