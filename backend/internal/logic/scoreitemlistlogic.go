// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScoreItemListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreItemListLogic {
	return &ScoreItemListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreItemListLogic) ScoreItemList(req *types.ScoreItemListReq) (resp *types.ScoreItemListResp, err error) {
	dim := int64(0)
	if req != nil {
		dim = req.DimensionId
	}
	items, err := l.svcCtx.ScoreItemRepo.List(l.ctx, dim)
	if err != nil {
		return nil, err
	}
	res := make([]types.ScoreItemResp, 0, len(items))
	for _, it := range items {
		res = append(res, types.ScoreItemResp{Id: it.ID, DimensionId: it.DimensionID, Name: it.Name, Score: it.Score})
	}
	return &types.ScoreItemListResp{Items: res}, nil
}
