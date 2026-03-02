package middleware

import (
	"net/http"

	"class-management-system/backend/internal/httperr"
)

func RequireAuth(allowPaths map[string]struct{}, authMw *AuthMiddleware) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowPaths[r.URL.Path]; ok {
				next(w, r)
				return
			}
			if authMw == nil {
				panic("auth middleware is nil")
			}
			authMw.Handle(func(w http.ResponseWriter, r *http.Request) {
				next(w, r)
			})(w, r)
		}
	}
}

func RequireAuthedCtx(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := UsernameFromCtx(r.Context())
		if !ok {
			panic(&httperr.Error{Code: http.StatusUnauthorized, Msg: "unauthorized"})
		}
		next(w, r)
	}
}
