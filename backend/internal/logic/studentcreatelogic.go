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

type StudentCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudentCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudentCreateLogic {
	return &StudentCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudentCreateLogic) StudentCreate(req *types.StudentCreateReq) (resp *types.StudentResp, err error) {
	if req == nil || req.StudentNo == "" || req.Name == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing required fields"}
	}
	s := &model.Student{
		StudentNo:  req.StudentNo,
		Name:       req.Name,
		Gender:     req.Gender,
		Phone:      req.Phone,
		Position:   req.Position,
		GroupID:    0,
		TotalScore: 0,
	}
	if err := l.svcCtx.StudentRepo.Create(l.ctx, s); err != nil {
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
