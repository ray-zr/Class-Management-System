package repository

import (
	"context"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
)

type GroupRepo struct {
	db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

func (r *GroupRepo) Create(ctx context.Context, g *model.Group) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *GroupRepo) List(ctx context.Context) ([]model.Group, error) {
	var res []model.Group
	if err := r.db.WithContext(ctx).Order("id asc").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *GroupRepo) Get(ctx context.Context, id int64) (*model.Group, error) {
	var g model.Group
	if err := r.db.WithContext(ctx).First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GroupRepo) UpdateName(ctx context.Context, id int64, name string) (*model.Group, error) {
	if err := r.db.WithContext(ctx).Model(&model.Group{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *GroupRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Student{}).Where("group_id = ?", id).Update("group_id", 0).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Group{}, id).Error
	})
}

func (r *GroupRepo) AvgScore(ctx context.Context, groupID int64) (int64, error) {
	var avg *float64
	if err := r.db.WithContext(ctx).Model(&model.Student{}).Where("group_id = ?", groupID).Select("avg(total_score)").Scan(&avg).Error; err != nil {
		return 0, err
	}
	if avg == nil {
		return 0, nil
	}
	return int64(*avg + 0.5), nil
}
