// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"class-management-system/backend/internal/logic"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func StudentUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt64(r, "id")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req types.StudentUpdateReq

		bodyBytes, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			httpx.ErrorCtx(r.Context(), w, readErr)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		provided := map[string]bool{}
		if len(bodyBytes) > 0 {
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(bodyBytes, &raw); err == nil {
				for k := range raw {
					provided[k] = true
				}
			}
		}

		l := logic.NewStudentUpdateLogic(r.Context(), svcCtx)
		resp, err := l.StudentUpdate(id, &req, provided)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
