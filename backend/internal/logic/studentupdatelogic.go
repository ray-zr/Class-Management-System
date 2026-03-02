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

type StudentUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudentUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudentUpdateLogic {
	return &StudentUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudentUpdateLogic) StudentUpdate(id int64, req *types.StudentUpdateReq, provided map[string]bool) (resp *types.StudentResp, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	if provided == nil {
		provided = map[string]bool{}
	}
	updates := map[string]any{}
	if provided["studentNo"] {
		updates["student_no"] = req.StudentNo
	}
	if provided["name"] {
		updates["name"] = req.Name
	}
	if provided["gender"] {
		updates["gender"] = req.Gender
	}
	if provided["phone"] {
		updates["phone"] = req.Phone
	}
	if provided["position"] {
		updates["position"] = req.Position
	}
	if provided["groupId"] {
		updates["group_id"] = req.GroupId
	}
	if len(updates) == 0 {
		s, err := l.svcCtx.StudentRepo.Get(l.ctx, id)
		if err != nil {
			return nil, err
		}
		return &types.StudentResp{
			Id:         s.ID,
			StudentNo:  s.StudentNo,
			Name:       s.Name,
			Gender:     s.Gender,
			Phone:      s.Phone,
			Position:   s.Position,
			GroupId:    s.GroupID,
			TotalScore: s.TotalScore,
			CreatedAt:  s.CreatedAt.Unix(),
			UpdatedAt:  s.UpdatedAt.Unix(),
		}, nil
	}
	s, err := l.svcCtx.StudentRepo.Update(l.ctx, id, updates)
	if err != nil {
		return nil, err
	}
	return &types.StudentResp{
		Id:         s.ID,
		StudentNo:  s.StudentNo,
		Name:       s.Name,
		Gender:     s.Gender,
		Phone:      s.Phone,
		Position:   s.Position,
		GroupId:    s.GroupID,
		TotalScore: s.TotalScore,
		CreatedAt:  s.CreatedAt.Unix(),
		UpdatedAt:  s.UpdatedAt.Unix(),
	}, nil
}
