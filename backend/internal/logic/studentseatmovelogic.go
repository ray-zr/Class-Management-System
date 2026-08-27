// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StudentSeatMoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudentSeatMoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudentSeatMoveLogic {
	return &StudentSeatMoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudentSeatMoveLogic) StudentSeatMove(req *types.StudentSeatMoveReq) (resp *types.Empty, err error) {
	if err := validateStudentSeatMove(req); err != nil {
		return nil, err
	}
	if req.GroupId > 0 {
		if _, err := l.svcCtx.GroupRepo.Get(l.ctx, req.GroupId); err != nil {
			return nil, err
		}
	}
	moving, err := l.svcCtx.StudentRepo.Get(l.ctx, req.StudentId)
	if err != nil {
		return nil, err
	}
	if req.TargetStudentId == req.StudentId {
		if moving.GroupID != req.GroupId {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "target student is not in target group"}
		}
		return &types.Empty{}, nil
	}
	if req.TargetStudentId > 0 {
		target, err := l.svcCtx.StudentRepo.Get(l.ctx, req.TargetStudentId)
		if err != nil {
			return nil, err
		}
		if target.GroupID != req.GroupId {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "target student is not in target group"}
		}
	}
	if err := l.svcCtx.StudentRepo.MoveSeat(l.ctx, req.StudentId, req.GroupId, req.TargetStudentId); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func validateStudentSeatMove(req *types.StudentSeatMoveReq) error {
	if req == nil || req.StudentId <= 0 {
		return &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid studentId"}
	}
	if req.GroupId < 0 {
		return &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid groupId"}
	}
	if req.TargetStudentId < 0 {
		return &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid targetStudentId"}
	}
	return nil
}
