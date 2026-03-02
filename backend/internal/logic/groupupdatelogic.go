// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUpdateLogic {
	return &GroupUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupUpdateLogic) GroupUpdate(id int64, req *types.GroupUpdateReq) (resp *types.GroupResp, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	if req == nil || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing name"}
	}
	g, err := l.svcCtx.GroupRepo.UpdateName(l.ctx, id, req.Name)
	if err != nil {
		return nil, err
	}
	avg, err := l.svcCtx.GroupRepo.AvgScore(l.ctx, g.ID)
	if err != nil {
		return nil, err
	}
	return &types.GroupResp{Id: g.ID, Name: g.Name, AvgScore: avg, CreatedAt: g.CreatedAt.Unix(), UpdatedAt: g.UpdatedAt.Unix()}, nil
}
