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
		if err := tx.First(&e, id).Error; err != nil {
			return err
		}
		res := tx.Model(&model.Student{}).
			Where("id = ?", e.StudentID).
			Update("total_score", gorm.Expr("total_score - ?", e.Score))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Delete(&model.ScoreEntry{}, id).Error
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		return nil, err
	}

	return &types.Empty{}, nil
}
