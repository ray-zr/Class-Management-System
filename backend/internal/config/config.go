// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Mysql struct {
		Dsn string
	}

	Auth struct {
		Username     string `json:",optional"`
		PasswordHash string `json:",optional"`
		JwtSecret    string `json:",optional"`
		JwtExpireSec int64
	}

	App struct {
		RecentScoreItemsN int64
		RankingTopN       int64
	}
}
