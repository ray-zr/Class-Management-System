package repository

import (
	"context"
	"errors"
	"sort"

	"class-management-system/backend/internal/model"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentRepo struct {
	db *gorm.DB
}

func NewStudentRepo(db *gorm.DB) *StudentRepo {
	return &StudentRepo{db: db}
}

func (r *StudentRepo) Create(ctx context.Context, s *model.Student) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		position, err := nextStudentSeatPosition(tx, s.GroupID)
		if err != nil {
			return err
		}
		s.SeatPosition = position
		return tx.Create(s).Error
	})
}

func (r *StudentRepo) CreateOrRestoreByStudentNo(ctx context.Context, s *model.Student) (*model.Student, error) {
	if s == nil || s.StudentNo == "" {
		return nil, gorm.ErrInvalidData
	}
	var restored *model.Student
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		position, err := nextStudentSeatPosition(tx, 0)
		if err != nil {
			return err
		}
		s.SeatPosition = position
		createErr := tx.Create(s).Error
		if createErr == nil {
			restored = s
			return nil
		}
		var me *mysql.MySQLError
		if !errors.As(createErr, &me) || me.Number != 1062 {
			return createErr
		}

		var existing model.Student
		if err := tx.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("student_no = ?", s.StudentNo).First(&existing).Error; err != nil {
			return createErr
		}
		if !existing.DeletedAt.Valid {
			return createErr
		}
		updates := map[string]any{
			"name":          s.Name,
			"gender":        s.Gender,
			"phone":         s.Phone,
			"position":      s.Position,
			"group_id":      0,
			"seat_position": position,
			"deleted_at":    nil,
		}
		if err := tx.Unscoped().Model(&model.Student{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return err
		}
		var current model.Student
		if err := tx.First(&current, existing.ID).Error; err != nil {
			return err
		}
		restored = &current
		return nil
	})
	return restored, err
}

func (r *StudentRepo) BatchUpsertByStudentNo(ctx context.Context, students []model.Student) error {
	if len(students) == 0 {
		return nil
	}
	for i := range students {
		if students[i].StudentNo == "" {
			return gorm.ErrInvalidData
		}
	}
	studentNos := make([]string, 0, len(students))
	for i := range students {
		studentNos = append(studentNos, students[i].StudentNo)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&model.Student{}).
			Where("student_no IN ? AND deleted_at IS NOT NULL", studentNos).
			Updates(map[string]any{"group_id": 0, "seat_position": 0, "deleted_at": nil}).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "student_no"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "gender", "phone", "position", "deleted_at"}),
		}).Create(&students).Error; err != nil {
			return err
		}
		return normalizeStudentSeatGroup(tx, 0)
	})
}

func (r *StudentRepo) Get(ctx context.Context, id int64) (*model.Student, error) {
	var s model.Student
	err := r.db.WithContext(ctx).First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) GetForUpdate(ctx context.Context, id int64) (*model.Student, error) {
	var student model.Student
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&student, id).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepo) ListForUpdate(ctx context.Context, groupID int64) ([]model.Student, error) {
	query := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.Student{})
	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	var students []model.Student
	if err := query.Order("id asc").Find(&students).Error; err != nil {
		return nil, err
	}
	return students, nil
}

func (r *StudentRepo) AddScore(ctx context.Context, id, score int64) error {
	return r.db.WithContext(ctx).Model(&model.Student{}).
		Where("id = ?", id).
		Update("total_score", gorm.Expr("total_score + ?", score)).Error
}

func (r *StudentRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var student model.Student
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&student, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&student).Error; err != nil {
			return err
		}
		return normalizeStudentSeatGroup(tx, student.GroupID)
	})
}

func (r *StudentRepo) Update(ctx context.Context, id int64, updates map[string]any) (*model.Student, error) {
	var out *model.Student
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.Student
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			return err
		}
		if groupID, ok := updates["group_id"].(int64); ok && groupID > 0 {
			var group model.Group
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, groupID).Error; err != nil {
				return err
			}
		}
		if groupID, ok := updates["group_id"].(int64); ok && groupID != current.GroupID {
			if err := normalizeStudentSeatGroup(tx, current.GroupID); err != nil {
				return err
			}
			if err := normalizeStudentSeatGroup(tx, groupID); err != nil {
				return err
			}
			position, err := nextStudentSeatPosition(tx, groupID)
			if err != nil {
				return err
			}
			updates["seat_position"] = position
		}
		result := tx.Model(&model.Student{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if groupID, ok := updates["group_id"].(int64); ok && groupID != current.GroupID {
			if err := normalizeStudentSeatGroup(tx, current.GroupID); err != nil {
				return err
			}
		}
		student, err := (&StudentRepo{db: tx}).Get(ctx, id)
		if err != nil {
			return err
		}
		out = student
		return nil
	})
	return out, err
}

func (r *StudentRepo) MoveSeat(ctx context.Context, studentID, groupID, targetStudentID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if groupID > 0 {
			var group model.Group
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, groupID).Error; err != nil {
				return err
			}
		}

		var moving model.Student
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&moving, studentID).Error; err != nil {
			return err
		}
		if targetStudentID == studentID {
			if moving.GroupID != groupID {
				return gorm.ErrInvalidData
			}
			return nil
		}

		groupIDs := []int64{moving.GroupID}
		if groupID != moving.GroupID {
			groupIDs = append(groupIDs, groupID)
		}
		sort.Slice(groupIDs, func(i, j int) bool { return groupIDs[i] < groupIDs[j] })
		for _, currentGroupID := range groupIDs {
			if err := normalizeStudentSeatGroup(tx, currentGroupID); err != nil {
				return err
			}
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&moving, studentID).Error; err != nil {
			return err
		}

		var target model.Student
		if targetStudentID > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&target, targetStudentID).Error; err != nil {
				return err
			}
			if target.GroupID != groupID {
				return gorm.ErrInvalidData
			}
		}

		if moving.GroupID == groupID {
			if targetStudentID > 0 {
				movingPosition := moving.SeatPosition
				if err := tx.Model(&model.Student{}).Where("id = ?", target.ID).Update("seat_position", movingPosition).Error; err != nil {
					return err
				}
				return tx.Model(&model.Student{}).Where("id = ?", moving.ID).Update("seat_position", target.SeatPosition).Error
			}
			var lastPosition int64
			if err := tx.Model(&model.Student{}).Where("group_id = ?", groupID).Select("COALESCE(MAX(seat_position), 0)").Scan(&lastPosition).Error; err != nil {
				return err
			}
			if moving.SeatPosition == lastPosition {
				return nil
			}
			if err := tx.Model(&model.Student{}).
				Where("group_id = ? AND seat_position > ?", groupID, moving.SeatPosition).
				Update("seat_position", gorm.Expr("seat_position - 1")).Error; err != nil {
				return err
			}
			return tx.Model(&model.Student{}).Where("id = ?", moving.ID).Update("seat_position", lastPosition).Error
		}

		if err := tx.Model(&model.Student{}).
			Where("group_id = ? AND seat_position > ?", moving.GroupID, moving.SeatPosition).
			Update("seat_position", gorm.Expr("seat_position - 1")).Error; err != nil {
			return err
		}

		insertPosition := target.SeatPosition
		if targetStudentID == 0 {
			if err := tx.Model(&model.Student{}).Where("group_id = ?", groupID).Select("COALESCE(MAX(seat_position), 0) + 1").Scan(&insertPosition).Error; err != nil {
				return err
			}
		} else if err := tx.Model(&model.Student{}).
			Where("group_id = ? AND seat_position >= ?", groupID, insertPosition).
			Update("seat_position", gorm.Expr("seat_position + 1")).Error; err != nil {
			return err
		}
		return tx.Model(&model.Student{}).Where("id = ?", moving.ID).Updates(map[string]any{
			"group_id":      groupID,
			"seat_position": insertPosition,
		}).Error
	})
}

func nextStudentSeatPosition(tx *gorm.DB, groupID int64) (int64, error) {
	var position int64
	if err := tx.Model(&model.Student{}).Where("group_id = ?", groupID).
		Select("COALESCE(MAX(seat_position), 0) + 1").Scan(&position).Error; err != nil {
		return 0, err
	}
	return position, nil
}

func normalizeStudentSeatGroup(tx *gorm.DB, groupID int64) error {
	var students []model.Student
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("group_id = ?", groupID).
		Order("CASE WHEN seat_position > 0 THEN 0 ELSE 1 END").
		Order("seat_position asc").
		Order("id asc").Find(&students).Error; err != nil {
		return err
	}
	for index, student := range students {
		position := int64(index + 1)
		if student.SeatPosition == position {
			continue
		}
		if err := tx.Model(&model.Student{}).Where("id = ?", student.ID).Update("seat_position", position).Error; err != nil {
			return err
		}
	}
	return nil
}

type StudentListFilter struct {
	Keyword string
	GroupID int64
	Offset  int64
	Limit   int64
}

func (r *StudentRepo) List(ctx context.Context, f StudentListFilter) (total int64, items []model.Student, err error) {
	q := r.db.WithContext(ctx).Model(&model.Student{})
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ? OR student_no LIKE ?", kw, kw)
	}
	if f.GroupID != 0 {
		q = q.Where("group_id = ?", f.GroupID)
	}
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if f.Limit > 0 {
		q = q.Offset(int(f.Offset)).Limit(int(f.Limit))
	}
	var res []model.Student
	if err := q.Order("id desc").Find(&res).Error; err != nil {
		return 0, nil, err
	}
	return total, res, nil
}
