package lib

import (
	"github.com/gin-gonic/gin"
	"mercury/x/ginx"
)

func registerMiddleware(engine *gin.Engine) {
	engine.NoMethod(ginx.NoMethodHandler())
	engine.NoRoute(ginx.NoRouteHandler())

	engine.Use(ginx.Recovery(), ginx.Logger(), ginx.CORS())
}
