// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"class-management-system/backend/internal/logic"
	"class-management-system/backend/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RollcallPickHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewRollcallPickLogic(r.Context(), svcCtx)
		resp, err := l.RollcallPick()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
