// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupListLogic {
	return &GroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupListLogic) GroupList() (resp *types.GroupListResp, err error) {
	gs, err := l.svcCtx.GroupRepo.List(l.ctx)
	if err != nil {
		return nil, err
	}
	items := make([]types.GroupResp, 0, len(gs))
	for _, g := range gs {
		avg, err := l.svcCtx.GroupRepo.AvgScore(l.ctx, g.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, types.GroupResp{Id: g.ID, Name: g.Name, AvgScore: avg, CreatedAt: g.CreatedAt.Unix(), UpdatedAt: g.UpdatedAt.Unix()})
	}
	return &types.GroupListResp{Items: items}, nil
}
