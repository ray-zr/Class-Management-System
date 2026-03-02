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

type StudentGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudentGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudentGetLogic {
	return &StudentGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudentGetLogic) StudentGet(id int64) (resp *types.StudentResp, err error) {
	if id <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid id"}
	}
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
