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

type DimensionDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDimensionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DimensionDeleteLogic {
	return &DimensionDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DimensionDeleteLogic) DimensionDelete(id int64) (resp *types.Empty, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	if _, err := l.svcCtx.DimensionRepo.Get(l.ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		return nil, err
	}
	if err := l.svcCtx.DimensionRepo.Delete(l.ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusNotFound, Msg: "not found"}
		}
		if errors.Is(err, repository.ErrDimensionHasScoreItems) {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "该维度下存在积分项，无法删除"}
		}
		if errors.Is(err, repository.ErrDimensionHasScoreEntries) {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "该维度下存在积分记录，无法删除"}
		}
		return nil, err
	}
	return &types.Empty{}, nil
}
