// @title           commentTree API
// @version         1.0
// @description     API для сервиса комментариев
// @BasePath        /

package main

import (
	"commentTree/internal/app"
	"commentTree/internal/config"
	"commentTree/internal/di"
	"commentTree/internal/storage/db"
	"commentTree/internal/web"
	wbzlog "github.com/wb-go/wbf/zlog"
	"go.uber.org/fx"
)

func main() {
	wbzlog.Init()
	app := fx.New(
		fx.Provide(
			config.NewAppConfig,
			db.NewPostgres,

			func(db *db.Postgres) app.DbProvider {
				return db
			},
			app.NewCommentService,

			func(service *app.CommentService) web.CommentService {
				return service
			},
			web.NewCommentHandler,
		),
		fx.Invoke(
			di.StartHTTPServer,
			di.ClosePostgresOnStop,
		),
	)

	app.Run()
}
