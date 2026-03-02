package middleware

import (
	"net/http"
	"strings"

	"class-management-system/backend/internal/auth"
	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAuthMiddleware(svcCtx *svc.ServiceContext) *AuthMiddleware {
	return &AuthMiddleware{svcCtx: svcCtx}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			httpx.Error(w, &httperr.Error{Code: http.StatusUnauthorized, Msg: "missing authorization"})
			return
		}
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.Error(w, &httperr.Error{Code: http.StatusUnauthorized, Msg: "invalid authorization"})
			return
		}
		claims, err := auth.Parse(parts[1], m.svcCtx.Config.Auth.JwtSecret)
		if err != nil {
			httpx.Error(w, &httperr.Error{Code: http.StatusUnauthorized, Msg: "invalid token"})
			return
		}
		r = r.WithContext(withUsername(r.Context(), claims.Username))
		next(w, r)
	}
}
