// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ScoreItemUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreItemUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreItemUpdateLogic {
	return &ScoreItemUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreItemUpdateLogic) ScoreItemUpdate(id int64, req *types.ScoreItemUpdateReq) (resp *types.ScoreItemResp, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	if req == nil || req.DimensionId <= 0 || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	updates := map[string]any{
		"dimension_id": req.DimensionId,
		"name":         req.Name,
		"score":        req.Score,
	}
	it, err := l.svcCtx.ScoreItemRepo.Update(l.ctx, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		return nil, err
	}
	return &types.ScoreItemResp{Id: it.ID, DimensionId: it.DimensionID, Name: it.Name, Score: it.Score}, nil
}
