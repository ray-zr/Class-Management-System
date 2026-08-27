// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"class-management-system/backend/internal/logic"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func StudentSeatMoveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StudentSeatMoveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewStudentSeatMoveLogic(r.Context(), svcCtx)
		resp, err := l.StudentSeatMove(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
