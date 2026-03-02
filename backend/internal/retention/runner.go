package retention

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type Runner struct {
	logx.Logger
	Cleaner  *Cleaner
	Locker   Locker
	LockName string
	Every    time.Duration
	KeepDays int64
}

func (r *Runner) Run(ctx context.Context) {
	if r.Cleaner == nil {
		return
	}
	if r.Every <= 0 {
		r.Every = time.Hour
	}
	if r.LockName == "" {
		r.LockName = "cms_retention_cleanup"
	}
	go func() {
		ticker := time.NewTicker(r.Every)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.Cleaner.CleanupWithLock(ctx, r.Locker, r.LockName, r.KeepDays); err != nil {
					r.Logger.Errorf("retention cleanup failed: %v", err)
				}
			}
		}
	}()
}
