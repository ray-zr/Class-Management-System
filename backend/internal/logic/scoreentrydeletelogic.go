// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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

func (l *ScoreEntryDeleteLogic) ScoreEntryDelete(id int64, req *types.ScoreEntryRevokeReq) (resp *types.Empty, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	reason := ""
	if req != nil {
		reason = strings.TrimSpace(req.Reason)
	}
	if reason == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "revoke reason is required"}
	}
	if utf8.RuneCountInString(reason) > 255 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "revoke reason is too long"}
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var e model.ScoreEntry
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, id).Error; err != nil {
			return err
		}
		if e.RevokedAt != nil {
			return &httperr.Error{Code: http.StatusConflict, Msg: "score entry was already revoked"}
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
		now := time.Now()
		result := tx.Model(&model.ScoreEntry{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{
			"revoked_at":    now,
			"revoke_reason": reason,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return &httperr.Error{Code: http.StatusConflict, Msg: "score entry was already revoked"}
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
