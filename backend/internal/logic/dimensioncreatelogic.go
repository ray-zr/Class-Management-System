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

type DimensionCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDimensionCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DimensionCreateLogic {
	return &DimensionCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DimensionCreateLogic) DimensionCreate(req *types.DimensionCreateReq) (resp *types.DimensionResp, err error) {
	if req == nil || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing name"}
	}
	d := &model.Dimension{Name: req.Name}
	if err := l.svcCtx.DimensionRepo.Create(l.ctx, d); err != nil {
		return nil, err
	}
	return &types.DimensionResp{Id: d.ID, Name: d.Name}, nil
}
