package repository

import (
	"context"
	"math/rand"
	"time"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
)

type RollcallRepo struct{ db *gorm.DB }

func NewRollcallRepo(db *gorm.DB) *RollcallRepo { return &RollcallRepo{db: db} }

func (r *RollcallRepo) StartRound(ctx context.Context, roundID string, fair bool) error {
	return r.db.WithContext(ctx).Create(&model.RollcallRound{RoundID: roundID, Fair: fair, Active: true}).Error
}

func (r *RollcallRepo) EndRound(ctx context.Context, roundID string) error {
	return r.db.WithContext(ctx).Model(&model.RollcallRound{}).Where("round_id = ?", roundID).Update("active", false).Error
}

func (r *RollcallRepo) RoundActive(ctx context.Context, roundID string) (bool, error) {
	var rr model.RollcallRound
	err := r.db.WithContext(ctx).Where("round_id = ?", roundID).First(&rr).Error
	if err != nil {
		return false, err
	}
	return rr.Active, nil
}

func (r *RollcallRepo) GetRound(ctx context.Context, roundID string) (*model.RollcallRound, error) {
	var rr model.RollcallRound
	if err := r.db.WithContext(ctx).Where("round_id = ?", roundID).First(&rr).Error; err != nil {
		return nil, err
	}
	return &rr, nil
}

func (r *RollcallRepo) Pick(ctx context.Context, roundID string, fair bool) (pickedStudentID int64, remaining int64, err error) {
	var allIDs []int64
	if err := r.db.WithContext(ctx).Table("students").Select("id").Scan(&allIDs).Error; err != nil {
		return 0, 0, err
	}
	if len(allIDs) == 0 {
		return 0, 0, gorm.ErrRecordNotFound
	}

	used := map[int64]struct{}{}
	if fair {
		var usedIDs []int64
		if err := r.db.WithContext(ctx).Model(&model.RollcallPicked{}).Where("round_id = ?", roundID).Select("student_id").Scan(&usedIDs).Error; err != nil {
			return 0, 0, err
		}
		for _, id := range usedIDs {
			used[id] = struct{}{}
		}
	}

	candidates := make([]int64, 0, len(allIDs))
	for _, id := range allIDs {
		if fair {
			if _, ok := used[id]; ok {
				continue
			}
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return 0, 0, gorm.ErrRecordNotFound
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rnd.Intn(len(candidates))
	picked := candidates[idx]

	if fair {
		if err := r.db.WithContext(ctx).Create(&model.RollcallPicked{RoundID: roundID, StudentID: picked}).Error; err != nil {
			return 0, 0, err
		}
		remaining = int64(len(candidates) - 1)
	} else {
		remaining = int64(len(allIDs) - 1)
	}
	return picked, remaining, nil
}

func (r *RollcallRepo) Reset(ctx context.Context, roundID string) error {
	if err := r.db.WithContext(ctx).Where("round_id = ?", roundID).Delete(&model.RollcallPicked{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.RollcallRound{}).Where("round_id = ?", roundID).Update("active", true).Error
}
