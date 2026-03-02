// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"
	"time"

	"class-management-system/backend/internal/auth"
	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	if req.Username != l.svcCtx.Config.Auth.Username {
		return nil, &httperr.Error{Code: http.StatusUnauthorized, Msg: "invalid credentials"}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(l.svcCtx.Config.Auth.PasswordHash), []byte(req.Password)); err != nil {
		return nil, &httperr.Error{Code: http.StatusUnauthorized, Msg: "invalid credentials"}
	}
	jwtToken, expAt, err := auth.Sign(req.Username, l.svcCtx.Config.Auth.JwtSecret, l.svcCtx.Config.Auth.JwtExpireSec, time.Now())
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{AccessToken: jwtToken, ExpireAt: expAt}, nil
}
