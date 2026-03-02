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

type DimensionUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDimensionUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DimensionUpdateLogic {
	return &DimensionUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DimensionUpdateLogic) DimensionUpdate(id int64, req *types.DimensionUpdateReq) (resp *types.DimensionResp, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	if req == nil || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing name"}
	}
	d, err := l.svcCtx.DimensionRepo.UpdateName(l.ctx, id, req.Name)
	if err != nil {
		return nil, err
	}
	return &types.DimensionResp{Id: d.ID, Name: d.Name}, nil
}
