package di

import (
	"commentTree/internal/config"
	"commentTree/internal/storage/db"
	"commentTree/internal/web"
	"context"
	"fmt"
	wbgin "github.com/wb-go/wbf/ginext"
	"go.uber.org/fx"
	"log"
	"net/http"
)

func StartHTTPServer(lc fx.Lifecycle, CommentHandler *web.CommentHandler, config *config.AppConfig) {
	router := wbgin.New(config.GinConfig.Mode)

	router.Use(wbgin.Logger(), wbgin.Recovery())
	router.Use(func(c *wbgin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	web.RegisterRoutes(router, CommentHandler)

	addres := fmt.Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.Port)
	server := &http.Server{
		Addr:    addres,
		Handler: router.Engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Server started")
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Printf("ListenAndServe error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Printf("Shutting down server...")
			return server.Close()
		},
	})
}

func ClosePostgresOnStop(lc fx.Lifecycle, postgres *db.Postgres) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("Closing Postgres connections...")
			if err := postgres.Close(); err != nil {
				log.Printf("Failed to close Postgres: %v", err)
				return err
			}
			log.Println("Postgres closed successfully")
			return nil
		},
	})
}
