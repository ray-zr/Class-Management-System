// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/logic"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		now := time.Now()
		if !svcCtx.LoginLimiter.Allow(key, now) {
			httpx.ErrorCtx(r.Context(), w, &httperr.Error{Code: http.StatusTooManyRequests, Msg: "too many login attempts"})
			return
		}
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.LoginLimiter.Failure(key, now)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			svcCtx.LoginLimiter.Failure(key, now)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			svcCtx.LoginLimiter.Success(key)
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
