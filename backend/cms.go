// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"class-management-system/backend/internal/config"
	"class-management-system/backend/internal/db"
	"class-management-system/backend/internal/handler"
	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/middleware"
	"class-management-system/backend/internal/svc"

	"github.com/go-sql-driver/mysql"
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
	if err := config.ApplyAuthEnv(&c, os.LookupEnv); err != nil {
		panic(fmt.Sprintf("invalid authentication configuration: %v", err))
	}
	httpx.SetErrorHandler(errorResponse)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	if err := db.PrepareCompatibilityMigrations(context.Background(), ctx.DB); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(context.Background(), ctx.DB); err != nil {
		panic(err)
	}
	if err := db.BackfillScoreEntrySnapshots(context.Background(), ctx.DB); err != nil {
		panic(err)
	}
	server.Use(middleware.RequireAuth(map[string]struct{}{
		"/api/health":     {},
		"/api/auth/login": {},
	}, middleware.NewAuthMiddleware(ctx)))
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func errorResponse(err error) (int, any) {
	if e, ok := err.(*httperr.Error); ok {
		return e.Code, map[string]any{"code": e.Code, "message": e.Msg}
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return http.StatusRequestEntityTooLarge, map[string]any{"code": http.StatusRequestEntityTooLarge, "message": "request body too large"}
	}
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var numberErr *strconv.NumError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.As(err, &numberErr) || errors.Is(err, io.ErrUnexpectedEOF) {
		return http.StatusBadRequest, map[string]any{"code": http.StatusBadRequest, "message": "invalid request"}
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return http.StatusConflict, map[string]any{"code": http.StatusConflict, "message": "record already exists"}
		case 1406:
			return http.StatusBadRequest, map[string]any{"code": http.StatusBadRequest, "message": "field value too long"}
		}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, map[string]any{"code": http.StatusNotFound, "message": "not found"}
	}
	if errors.Is(err, gorm.ErrInvalidData) {
		return http.StatusBadRequest, map[string]any{"code": http.StatusBadRequest, "message": "invalid request"}
	}
	logx.Errorf("request failed: %v", err)
	return http.StatusInternalServerError, map[string]any{"code": http.StatusInternalServerError, "message": "internal server error"}
}
