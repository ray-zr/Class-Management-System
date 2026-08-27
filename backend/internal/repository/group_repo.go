package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GroupRepo struct {
	db *gorm.DB
}

type GroupLayoutPosition struct {
	ID       int64
	Position int64
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

func (r *GroupRepo) Create(ctx context.Context, g *model.Group) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var last model.Group
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Order("layout_position desc").
			Order("id desc").
			First(&last).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		g.LayoutPosition = last.LayoutPosition + 1
		return tx.Create(g).Error
	})
}

func (r *GroupRepo) List(ctx context.Context) ([]model.Group, error) {
	var res []model.Group
	if err := r.db.WithContext(ctx).Order("layout_position asc").Order("id asc").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *GroupRepo) UpdateLayout(ctx context.Context, positions []GroupLayoutPosition) error {
	if len(positions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := make([]int64, 0, len(positions))
		for _, item := range positions {
			ids = append(ids, item.ID)
		}
		var groups []model.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Find(&groups).Error; err != nil {
			return err
		}
		if len(groups) != len(positions) {
			return gorm.ErrRecordNotFound
		}
		for _, item := range positions {
			if err := tx.Model(&model.Group{}).Where("id = ?", item.ID).Update("layout_position", item.Position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GroupRepo) Get(ctx context.Context, id int64) (*model.Group, error) {
	var g model.Group
	if err := r.db.WithContext(ctx).First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GroupRepo) GetForUpdate(ctx context.Context, id int64) (*model.Group, error) {
	var group model.Group
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepo) UpdateName(ctx context.Context, id int64, name string) (*model.Group, error) {
	if err := r.db.WithContext(ctx).Model(&model.Group{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *GroupRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var group model.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Student{}).Where("group_id = ?", id).Updates(map[string]any{"group_id": 0, "seat_position": 0}).Error; err != nil {
			return err
		}
		if err := normalizeStudentSeatGroup(tx, 0); err != nil {
			return err
		}
		if err := tx.Delete(&group).Error; err != nil {
			return err
		}
		return tx.Model(&model.Group{}).
			Where("layout_position > ?", group.LayoutPosition).
			Update("layout_position", gorm.Expr("layout_position - 1")).Error
	})
}

func (r *GroupRepo) AvgScore(ctx context.Context, groupID int64) (float64, error) {
	var avg sql.NullFloat64
	if err := r.db.WithContext(ctx).
		Model(&model.Student{}).
		Where("group_id = ?", groupID).
		Select("avg(total_score)").
		Scan(&avg).Error; err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return math.Round(avg.Float64*100) / 100, nil
}
