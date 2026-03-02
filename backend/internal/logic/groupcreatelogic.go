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

type GroupCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupCreateLogic {
	return &GroupCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupCreateLogic) GroupCreate(req *types.GroupCreateReq) (resp *types.GroupResp, err error) {
	if req == nil || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing name"}
	}
	g := &model.Group{Name: req.Name}
	if err := l.svcCtx.GroupRepo.Create(l.ctx, g); err != nil {
		return nil, err
	}
	return &types.GroupResp{Id: g.ID, Name: g.Name, AvgScore: 0, CreatedAt: g.CreatedAt.Unix(), UpdatedAt: g.UpdatedAt.Unix()}, nil
}
