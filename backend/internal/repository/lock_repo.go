package repository

import (
	"context"

	"gorm.io/gorm"
)

type LockRepo struct{ db *gorm.DB }

func NewLockRepo(db *gorm.DB) *LockRepo { return &LockRepo{db: db} }

func (r *LockRepo) TryGetLock(ctx context.Context, name string) (bool, error) {
	var ok int
	if err := r.db.WithContext(ctx).Raw("SELECT GET_LOCK(?, 0)", name).Scan(&ok).Error; err != nil {
		return false, err
	}
	return ok == 1, nil
}

func (r *LockRepo) ReleaseLock(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Exec("SELECT RELEASE_LOCK(?)", name).Error
}
