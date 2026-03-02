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

func (r *StudentRepo) UpsertByStudentNo(ctx context.Context, s *model.Student) error {
	if s == nil || s.StudentNo == "" {
		return gorm.ErrInvalidData
	}
	if err := r.Create(ctx, s); err == nil {
		return nil
	} else {
		var me *mysql.MySQLError
		if !errors.As(err, &me) || me.Number != 1062 {
			return err
		}
		updates := map[string]any{
			"name":     s.Name,
			"gender":   s.Gender,
			"phone":    s.Phone,
			"position": s.Position,
		}
		return r.db.WithContext(ctx).Model(&model.Student{}).Where("student_no = ?", s.StudentNo).Updates(updates).Error
	}
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
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "student_no"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "gender", "phone", "position"}),
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

func (r *StudentRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Student{}, id).Error
}

func (r *StudentRepo) Update(ctx context.Context, id int64, updates map[string]any) (*model.Student, error) {
	if err := r.db.WithContext(ctx).Model(&model.Student{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
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
