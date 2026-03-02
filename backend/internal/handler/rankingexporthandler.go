// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"class-management-system/backend/internal/logic"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RankingExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RankingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewRankingExportLogic(r.Context(), svcCtx)
		file, err := l.RankingExport(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		month := req.Month
		if !req.Total && month == "" {
			month = time.Now().Format("2006-01")
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		filename := "总分积分排名汇总表.xlsx"
		if !req.Total {
			filename = "月度积分排名汇总表.xlsx"
		}
		w.Header().Set("Content-Disposition", contentDisposition(filename))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(file)
	}
}

func contentDisposition(filename string) string {
	fallback := "ranking.xlsx"
	q := url.QueryEscape(filename)
	q = strings.ReplaceAll(q, "+", "%20")
	return fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fallback, q)
}
