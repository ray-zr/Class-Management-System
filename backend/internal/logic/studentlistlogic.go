// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StudentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudentListLogic {
	return &StudentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudentListLogic) StudentList(req *types.StudentListReq) (resp *types.StudentListResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	offset := (page - 1) * size
	limit := size
	total, items, err := l.svcCtx.StudentRepo.List(l.ctx, repository.StudentListFilter{
		Keyword: req.Keyword,
		GroupID: req.GroupId,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}
	respItems := make([]types.StudentResp, 0, len(items))
	for _, s := range items {
		respItems = append(respItems, types.StudentResp{
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
		})
	}
	return &types.StudentListResp{Total: total, Items: respItems}, nil
}
