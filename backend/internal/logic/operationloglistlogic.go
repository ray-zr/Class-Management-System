// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"net/http"
	"strings"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperationLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationLogListLogic {
	return &OperationLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OperationLogListLogic) OperationLogList(req *types.OperationLogListReq) (resp *types.OperationLogListResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 || size > 200 {
		size = 20
	}
	start, end, err := operationLogRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	total, entries, err := l.svcCtx.ScoreEntryRepo.ListOperationLogs(l.ctx, repository.OperationLogListFilter{
		StudentID: req.StudentId,
		GroupID:   req.GroupId,
		Start:     start,
		End:       end,
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.ScoreEntryResp, 0, len(entries))
	for i := range entries {
		items = append(items, *scoreEntryResp(&entries[i]))
	}
	return &types.OperationLogListResp{Total: total, Items: items}, nil
}

func operationLogRange(startDate, endDate string) (time.Time, time.Time, error) {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		return time.Time{}, time.Time{}, nil
	}
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, &httperr.Error{Code: http.StatusBadRequest, Msg: "startDate and endDate must be provided together"}
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid startDate"}
	}
	endDay, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid endDate"}
	}
	if endDay.Before(start) {
		return time.Time{}, time.Time{}, &httperr.Error{Code: http.StatusBadRequest, Msg: "endDate must not be earlier than startDate"}
	}
	return start, endDay.AddDate(0, 0, 1), nil
}
