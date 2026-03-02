// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"class-management-system/backend/internal/logic"
	"class-management-system/backend/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func StudentImportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		l := logic.NewStudentImportLogic(r.Context(), svcCtx)
		resp, err := l.StudentImport(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
