package repository

import (
	"context"
	"errors"
	"time"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		var dimension model.Dimension
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&dimension, id).Error; err != nil {
			return err
		}
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
		return tx.Delete(&dimension).Error
	})
}

type ScoreItemRepo struct{ db *gorm.DB }

func NewScoreItemRepo(db *gorm.DB) *ScoreItemRepo { return &ScoreItemRepo{db: db} }

func (r *ScoreItemRepo) Create(ctx context.Context, it *model.ScoreItem) error {
	if it == nil || it.DimensionID <= 0 {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dimension model.Dimension
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&dimension, it.DimensionID).Error; err != nil {
			return err
		}
		return tx.Create(it).Error
	})
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

func (r *ScoreItemRepo) GetForUpdate(ctx context.Context, id int64) (*model.ScoreItem, error) {
	var item model.ScoreItem
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ScoreItemRepo) Update(ctx context.Context, id int64, updates map[string]any) (*model.ScoreItem, error) {
	var out *model.ScoreItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var before model.ScoreItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&before, id).Error; err != nil {
			return err
		}
		if dimensionID, ok := updates["dimension_id"].(int64); ok {
			var dimension model.Dimension
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&dimension, dimensionID).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.ScoreItem{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		after, err := (&ScoreItemRepo{db: tx}).Get(ctx, id)
		if err != nil {
			return err
		}
		out = after
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ScoreItemRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.ScoreItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return err
		}
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
		return tx.Delete(&item).Error
	})
}

type RecentScoreItemRepo struct{ db *gorm.DB }

func NewRecentScoreItemRepo(db *gorm.DB) *RecentScoreItemRepo { return &RecentScoreItemRepo{db: db} }

func (r *RecentScoreItemRepo) Touch(ctx context.Context, scoreItemID int64, usedAt time.Time) error {
	item := model.RecentScoreItem{ScoreItemID: scoreItemID, UsedAtUnix: usedAt.Unix()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "score_item_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"used_at_unix", "updated_at"}),
	}).Create(&item).Error
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

func (r *ScoreEntryRepo) Get(ctx context.Context, id int64) (*model.ScoreEntry, error) {
	var e model.ScoreEntry
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ScoreEntryRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.ScoreEntry{}, id).Error
}

type ScoreOperationRepo struct{ db *gorm.DB }

func NewScoreOperationRepo(db *gorm.DB) *ScoreOperationRepo { return &ScoreOperationRepo{db: db} }

func (r *ScoreOperationRepo) CreateOrGet(ctx context.Context, operation *model.ScoreOperation) (bool, *model.ScoreOperation, error) {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(operation)
	if result.Error != nil {
		return false, nil, result.Error
	}
	if result.RowsAffected == 1 {
		return true, operation, nil
	}
	var existing model.ScoreOperation
	if err := r.db.WithContext(ctx).Where("request_id = ?", operation.RequestID).First(&existing).Error; err != nil {
		return false, nil, err
	}
	return false, &existing, nil
}

func (r *ScoreOperationRepo) Complete(ctx context.Context, id, lastEntryID int64) error {
	result := r.db.WithContext(ctx).Model(&model.ScoreOperation{}).
		Where("id = ? AND last_entry_id = 0", id).
		Update("last_entry_id", lastEntryID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrInvalidData
	}
	return nil
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
	if !f.Since.IsZero() {
		q = q.Where("created_at >= ?", f.Since)
	}
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
