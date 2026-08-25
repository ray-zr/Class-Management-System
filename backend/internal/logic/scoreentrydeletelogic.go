// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScoreEntryDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreEntryDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreEntryDeleteLogic {
	return &ScoreEntryDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreEntryDeleteLogic) ScoreEntryDelete(id int64) (resp *types.Empty, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var e model.ScoreEntry
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, id).Error; err != nil {
			return err
		}
		var student model.Student
		studentErr := tx.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).First(&student, e.StudentID).Error
		if studentErr == nil {
			if err := tx.Unscoped().Model(&model.Student{}).
				Where("id = ?", e.StudentID).
				Update("total_score", gorm.Expr("total_score - ?", e.Score)).Error; err != nil {
				return err
			}
		} else if !errors.Is(studentErr, gorm.ErrRecordNotFound) {
			return studentErr
		}
		result := tx.Delete(&model.ScoreEntry{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		return nil, err
	}

	return &types.Empty{}, nil
}
