package repository

import (
	"context"
	"math/rand/v2"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RollcallRepo struct{ db *gorm.DB }

func NewRollcallRepo(db *gorm.DB) *RollcallRepo { return &RollcallRepo{db: db} }

func (r *RollcallRepo) StartRound(ctx context.Context, roundID string, fair bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.RollcallRound{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}
		return tx.Create(&model.RollcallRound{RoundID: roundID, Fair: fair, Active: true}).Error
	})
}

func (r *RollcallRepo) EndRound(ctx context.Context, roundID string) error {
	return r.db.WithContext(ctx).Model(&model.RollcallRound{}).Where("round_id = ?", roundID).Update("active", false).Error
}

func (r *RollcallRepo) ActiveRound(ctx context.Context) (*model.RollcallRound, error) {
	var round model.RollcallRound
	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("id desc").First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *RollcallRepo) Pick(ctx context.Context, roundID string) (pickedStudentID int64, remaining int64, err error) {
	exhausted := false
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var round model.RollcallRound
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("round_id = ?", roundID).First(&round).Error; err != nil {
			return err
		}
		if !round.Active {
			return gorm.ErrRecordNotFound
		}

		var allIDs []int64
		if err := tx.Table("students").Select("id").Where("deleted_at IS NULL").Order("id asc").Scan(&allIDs).Error; err != nil {
			return err
		}
		if len(allIDs) == 0 {
			return gorm.ErrRecordNotFound
		}

		used := make(map[int64]struct{}, len(allIDs))
		if round.Fair {
			var usedIDs []int64
			if err := tx.Model(&model.RollcallPicked{}).Where("round_id = ?", roundID).Select("student_id").Scan(&usedIDs).Error; err != nil {
				return err
			}
			for _, id := range usedIDs {
				used[id] = struct{}{}
			}
		}

		candidates := make([]int64, 0, len(allIDs))
		for _, id := range allIDs {
			if _, alreadyPicked := used[id]; round.Fair && alreadyPicked {
				continue
			}
			candidates = append(candidates, id)
		}
		if len(candidates) == 0 {
			exhausted = true
			return tx.Model(&model.RollcallRound{}).Where("active = ?", true).Update("active", false).Error
		}

		pickedStudentID = candidates[rand.IntN(len(candidates))]
		if round.Fair {
			if err := tx.Create(&model.RollcallPicked{RoundID: roundID, StudentID: pickedStudentID}).Error; err != nil {
				return err
			}
			remaining = int64(len(candidates) - 1)
			if remaining == 0 {
				if err := tx.Model(&model.RollcallRound{}).Where("active = ?", true).Update("active", false).Error; err != nil {
					return err
				}
			}
		} else {
			remaining = int64(len(allIDs) - 1)
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	if exhausted {
		return 0, 0, gorm.ErrRecordNotFound
	}
	return pickedStudentID, remaining, nil
}

func (r *RollcallRepo) Reset(ctx context.Context, roundID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var round model.RollcallRound
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("round_id = ?", roundID).First(&round).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.RollcallRound{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}
		if err := tx.Where("round_id = ?", roundID).Delete(&model.RollcallPicked{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.RollcallRound{}).Where("id = ?", round.ID).Update("active", true).Error
	})
}
