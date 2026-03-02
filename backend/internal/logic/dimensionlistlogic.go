// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DimensionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDimensionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DimensionListLogic {
	return &DimensionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DimensionListLogic) DimensionList() (resp *types.DimensionListResp, err error) {
	items, err := l.svcCtx.DimensionRepo.List(l.ctx)
	if err != nil {
		return nil, err
	}
	res := make([]types.DimensionResp, 0, len(items))
	for _, d := range items {
		res = append(res, types.DimensionResp{Id: d.ID, Name: d.Name})
	}
	return &types.DimensionListResp{Items: res}, nil
}
