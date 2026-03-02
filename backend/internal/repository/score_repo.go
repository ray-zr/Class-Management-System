package repository

import (
	"context"
	"errors"
	"time"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
)

var ErrDimensionHasScoreItems = errors.New("dimension has score items")
var ErrDimensionHasScoreEntries = errors.New("dimension has score entries")
var ErrScoreItemHasScoreEntries = errors.New("score item has score entries")

type DimensionRepo struct{ db *gorm.DB }

func NewDimensionRepo(db *gorm.DB) *DimensionRepo { return &DimensionRepo{db: db} }

func (r *DimensionRepo) Create(ctx context.Context, d *model.Dimension) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *DimensionRepo) List(ctx context.Context) ([]model.Dimension, error) {
	var res []model.Dimension
	if err := r.db.WithContext(ctx).Order("id asc").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *DimensionRepo) Get(ctx context.Context, id int64) (*model.Dimension, error) {
	var d model.Dimension
	if err := r.db.WithContext(ctx).First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DimensionRepo) UpdateName(ctx context.Context, id int64, name string) (*model.Dimension, error) {
	if err := r.db.WithContext(ctx).Model(&model.Dimension{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *DimensionRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cnt int64
		if err := tx.Model(&model.ScoreItem{}).Where("dimension_id = ?", id).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return ErrDimensionHasScoreItems
		}
		if err := tx.Model(&model.ScoreEntry{}).Where("dimension_id = ?", id).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return ErrDimensionHasScoreEntries
		}
		return tx.Delete(&model.Dimension{}, id).Error
	})
}

type ScoreItemRepo struct{ db *gorm.DB }

func NewScoreItemRepo(db *gorm.DB) *ScoreItemRepo { return &ScoreItemRepo{db: db} }

func (r *ScoreItemRepo) Create(ctx context.Context, it *model.ScoreItem) error {
	return r.db.WithContext(ctx).Create(it).Error
}

func (r *ScoreItemRepo) List(ctx context.Context, dimensionID int64) ([]model.ScoreItem, error) {
	q := r.db.WithContext(ctx).Model(&model.ScoreItem{})
	if dimensionID != 0 {
		q = q.Where("dimension_id = ?", dimensionID)
	}
	var res []model.ScoreItem
	if err := q.Order("id desc").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ScoreItemRepo) Get(ctx context.Context, id int64) (*model.ScoreItem, error) {
	var it model.ScoreItem
	if err := r.db.WithContext(ctx).First(&it, id).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *ScoreItemRepo) Update(ctx context.Context, id int64, updates map[string]any) (*model.ScoreItem, error) {
	if err := r.db.WithContext(ctx).Model(&model.ScoreItem{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *ScoreItemRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cnt int64
		if err := tx.Model(&model.ScoreEntry{}).Where("score_item_id = ?", id).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return ErrScoreItemHasScoreEntries
		}
		if err := tx.Where("score_item_id = ?", id).Delete(&model.RecentScoreItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ScoreItem{}, id).Error
	})
}

type RecentScoreItemRepo struct{ db *gorm.DB }

func NewRecentScoreItemRepo(db *gorm.DB) *RecentScoreItemRepo { return &RecentScoreItemRepo{db: db} }

func (r *RecentScoreItemRepo) Touch(ctx context.Context, scoreItemID int64, usedAt time.Time) error {
	unix := usedAt.Unix()
	var existing model.RecentScoreItem
	err := r.db.WithContext(ctx).Where("score_item_id = ?", scoreItemID).First(&existing).Error
	if err == nil {
		return r.db.WithContext(ctx).Model(&model.RecentScoreItem{}).Where("score_item_id = ?", scoreItemID).Update("used_at_unix", unix).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.WithContext(ctx).Create(&model.RecentScoreItem{ScoreItemID: scoreItemID, UsedAtUnix: unix}).Error
}

func (r *RecentScoreItemRepo) ListRecent(ctx context.Context, limit int64) ([]int64, error) {
	if limit <= 0 {
		limit = 10
	}
	var ids []int64
	if err := r.db.WithContext(ctx).Model(&model.RecentScoreItem{}).Select("score_item_id").Order("used_at_unix desc").Limit(int(limit)).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

type ScoreEntryRepo struct{ db *gorm.DB }

func NewScoreEntryRepo(db *gorm.DB) *ScoreEntryRepo { return &ScoreEntryRepo{db: db} }

func (r *ScoreEntryRepo) Create(ctx context.Context, e *model.ScoreEntry) error {
	return r.db.WithContext(ctx).Create(e).Error
}

type ScoreEntryListFilter struct {
	StudentID int64
	GroupID   int64
	Since     time.Time
	Offset    int64
	Limit     int64
}

func (r *ScoreEntryRepo) List(ctx context.Context, f ScoreEntryListFilter) (total int64, items []model.ScoreEntry, err error) {
	q := r.db.WithContext(ctx).Model(&model.ScoreEntry{})
	if f.StudentID != 0 {
		q = q.Where("student_id = ?", f.StudentID)
	}
	if f.GroupID != 0 {
		q = q.Where("group_id = ?", f.GroupID)
	}
	q = q.Where("created_at >= ?", f.Since)
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if f.Limit > 0 {
		q = q.Offset(int(f.Offset)).Limit(int(f.Limit))
	}
	var res []model.ScoreEntry
	if err := q.Order("id desc").Find(&res).Error; err != nil {
		return 0, nil, err
	}
	return total, res, nil
}

func (r *ScoreEntryRepo) DeleteBefore(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&model.ScoreEntry{}).Error
}
