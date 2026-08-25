package auth

import (
	"sync"
	"time"
)

type loginAttempt struct {
	failures     int
	blockedUntil time.Time
}

type LoginLimiter struct {
	mu          sync.Mutex
	attempts    map[string]loginAttempt
	maxFailures int
	blockFor    time.Duration
}

func NewLoginLimiter(maxFailures int, blockFor time.Duration) *LoginLimiter {
	if maxFailures < 1 {
		maxFailures = 5
	}
	if blockFor <= 0 {
		blockFor = 5 * time.Minute
	}
	return &LoginLimiter{
		attempts:    make(map[string]loginAttempt),
		maxFailures: maxFailures,
		blockFor:    blockFor,
	}
}

func (l *LoginLimiter) Allow(key string, now time.Time) bool {
	if l == nil || key == "" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt := l.attempts[key]
	if !attempt.blockedUntil.IsZero() && now.Before(attempt.blockedUntil) {
		return false
	}
	if !attempt.blockedUntil.IsZero() {
		delete(l.attempts, key)
	}
	return true
}

func (l *LoginLimiter) Failure(key string, now time.Time) {
	if l == nil || key == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt := l.attempts[key]
	attempt.failures++
	if attempt.failures >= l.maxFailures {
		attempt.blockedUntil = now.Add(l.blockFor)
	}
	l.attempts[key] = attempt
}

func (l *LoginLimiter) Success(key string) {
	if l == nil || key == "" {
		return
	}
	l.mu.Lock()
	delete(l.attempts, key)
	l.mu.Unlock()
}
