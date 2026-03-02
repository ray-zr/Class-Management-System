// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScoreItemCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreItemCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreItemCreateLogic {
	return &ScoreItemCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreItemCreateLogic) ScoreItemCreate(req *types.ScoreItemCreateReq) (resp *types.ScoreItemResp, err error) {
	if req == nil || req.DimensionId <= 0 || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	it := &model.ScoreItem{DimensionID: req.DimensionId, Name: req.Name, Score: req.Score}
	if err := l.svcCtx.ScoreItemRepo.Create(l.ctx, it); err != nil {
		return nil, err
	}
	return &types.ScoreItemResp{Id: it.ID, DimensionId: it.DimensionID, Name: it.Name, Score: it.Score}, nil
}
