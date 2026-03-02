// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"class-management-system/backend/internal/config"
	"class-management-system/backend/internal/db"
	"class-management-system/backend/internal/handler"
	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/middleware"
	"class-management-system/backend/internal/retention"
	"class-management-system/backend/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"gorm.io/gorm"
)

var configFile = flag.String("f", "etc/cms-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	httpx.SetErrorHandler(func(err error) (int, any) {
		switch e := err.(type) {
		case *httperr.Error:
			return e.Code, map[string]any{"code": e.Code, "message": e.Msg}
		default:
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return http.StatusNotFound, map[string]any{"code": http.StatusNotFound, "message": "not found"}
			}
			if errors.Is(err, gorm.ErrInvalidData) {
				return http.StatusBadRequest, map[string]any{"code": http.StatusBadRequest, "message": "invalid request"}
			}
			return 500, map[string]any{"code": 500, "message": err.Error()}
		}
	})

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	if err := db.AutoMigrate(context.Background(), ctx.DB); err != nil {
		panic(err)
	}
	server.Use(middleware.RequireAuth(map[string]struct{}{
		"/api/health":     {},
		"/api/auth/login": {},
	}, middleware.NewAuthMiddleware(ctx)))
	(&retention.Runner{
		Logger:   logx.WithContext(context.Background()),
		Cleaner:  &retention.Cleaner{Logger: logx.WithContext(context.Background()), Repo: ctx.ScoreEntryRepo},
		Locker:   ctx.LockRepo,
		LockName: "cms_retention_cleanup",
		Every:    time.Duration(c.Retention.CleanupEverySec) * time.Second,
		KeepDays: c.Retention.ScoreEntryDays,
	}).Run(context.Background())
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
