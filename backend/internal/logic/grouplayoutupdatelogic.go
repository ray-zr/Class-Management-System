// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupLayoutUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupLayoutUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupLayoutUpdateLogic {
	return &GroupLayoutUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupLayoutUpdateLogic) GroupLayoutUpdate(req *types.GroupLayoutUpdateReq) (resp *types.Empty, err error) {
	groups, err := l.svcCtx.GroupRepo.List(l.ctx)
	if err != nil {
		return nil, err
	}
	positions, err := validateGroupLayout(req, groups)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.GroupRepo.UpdateLayout(l.ctx, positions); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func validateGroupLayout(req *types.GroupLayoutUpdateReq, groups []model.Group) ([]repository.GroupLayoutPosition, error) {
	if req == nil || len(req.Items) != len(groups) {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "layout must include every group"}
	}
	existing := make(map[int64]struct{}, len(groups))
	for _, group := range groups {
		existing[group.ID] = struct{}{}
	}
	seenIDs := make(map[int64]struct{}, len(req.Items))
	seenPositions := make(map[int64]struct{}, len(req.Items))
	positions := make([]repository.GroupLayoutPosition, 0, len(req.Items))
	for _, item := range req.Items {
		if _, ok := existing[item.Id]; !ok {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "layout contains unknown group"}
		}
		if item.Position < 1 || item.Position > int64(len(groups)) {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid layout position"}
		}
		if _, ok := seenIDs[item.Id]; ok {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "layout contains duplicate group"}
		}
		if _, ok := seenPositions[item.Position]; ok {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "layout contains duplicate position"}
		}
		seenIDs[item.Id] = struct{}{}
		seenPositions[item.Position] = struct{}{}
		positions = append(positions, repository.GroupLayoutPosition{ID: item.Id, Position: item.Position})
	}
	return positions, nil
}
