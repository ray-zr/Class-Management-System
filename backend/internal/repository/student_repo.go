package repository

import (
	"context"
	"errors"

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
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *StudentRepo) CreateOrRestoreByStudentNo(ctx context.Context, s *model.Student) (*model.Student, error) {
	if s == nil || s.StudentNo == "" {
		return nil, gorm.ErrInvalidData
	}
	var restored *model.Student
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			"name":       s.Name,
			"gender":     s.Gender,
			"phone":      s.Phone,
			"position":   s.Position,
			"group_id":   0,
			"deleted_at": nil,
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
	if err := r.db.WithContext(ctx).Unscoped().Model(&model.Student{}).
		Where("student_no IN ? AND deleted_at IS NOT NULL", studentNos).
		Updates(map[string]any{"group_id": 0, "deleted_at": nil}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "student_no"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "gender", "phone", "position", "deleted_at"}),
		}).
		Create(&students).Error
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
	result := r.db.WithContext(ctx).Delete(&model.Student{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *StudentRepo) Update(ctx context.Context, id int64, updates map[string]any) (*model.Student, error) {
	var out *model.Student
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if groupID, ok := updates["group_id"].(int64); ok && groupID > 0 {
			var group model.Group
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, groupID).Error; err != nil {
				return err
			}
		}
		result := tx.Model(&model.Student{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return result.Error
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
