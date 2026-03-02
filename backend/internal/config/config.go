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
		Username     string
		PasswordHash string
		JwtSecret    string
		JwtExpireSec int64
	}

	App struct {
		RecentScoreItemsN int64
		RankingTopN       int64
	}

	Retention struct {
		ScoreEntryDays  int64
		CleanupEverySec int64
	}
}
