package web

import (
	_ "commentTree/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	wbgin "github.com/wb-go/wbf/ginext"
)

func RegisterRoutes(engine *wbgin.Engine, handler *CommentHandler) {
	api := engine.Group("/api")
	{
		api.POST("/comments", handler.CreateComment)
		api.GET("/comments", handler.GetComments)
		api.DELETE("/comments/:id", handler.DeleteComments)
		api.GET("/swagger/*any", func(c *wbgin.Context) {
			httpSwagger.WrapHandler(c.Writer, c.Request)
		})
	}
}
