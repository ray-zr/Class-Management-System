package retention

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScoreEntryDeleter interface {
	DeleteBefore(ctx context.Context, before time.Time) error
}

type Cleaner struct {
	logx.Logger
	Repo ScoreEntryDeleter
}

type Locker interface {
	TryGetLock(ctx context.Context, name string) (bool, error)
	ReleaseLock(ctx context.Context, name string) error
}

func (c *Cleaner) Cleanup(ctx context.Context, keepDays int64) error {
	if c.Repo == nil {
		return nil
	}
	if keepDays <= 0 {
		keepDays = 30
	}
	before := time.Now().Add(-time.Duration(keepDays) * 24 * time.Hour)
	return c.Repo.DeleteBefore(ctx, before)
}

func (c *Cleaner) CleanupWithLock(ctx context.Context, lock Locker, lockName string, keepDays int64) error {
	if lock == nil {
		return c.Cleanup(ctx, keepDays)
	}
	ok, err := lock.TryGetLock(ctx, lockName)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	defer func() { _ = lock.ReleaseLock(ctx, lockName) }()
	return c.Cleanup(ctx, keepDays)
}
