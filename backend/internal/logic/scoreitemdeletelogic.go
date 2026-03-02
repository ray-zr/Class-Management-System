// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ScoreItemDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreItemDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreItemDeleteLogic {
	return &ScoreItemDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreItemDeleteLogic) ScoreItemDelete(id int64) (resp *types.Empty, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	if _, err := l.svcCtx.ScoreItemRepo.Get(l.ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		return nil, err
	}
	if err := l.svcCtx.ScoreItemRepo.Delete(l.ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		if errors.Is(err, repository.ErrScoreItemHasScoreEntries) {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "该积分项已被使用，无法删除"}
		}
		return nil, err
	}
	return &types.Empty{}, nil
}
