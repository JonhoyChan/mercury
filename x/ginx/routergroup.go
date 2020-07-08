package ginx

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	group *gin.RouterGroup
}

func NewGroup(group *gin.RouterGroup) *RouterGroup {
	g := &RouterGroup{group: group}
	g.GET("/metrics", Monitor())
	return g
}

func handlerFuncToGinHandlerFunc(h1 ...HandlerFunc) []gin.HandlerFunc {
	var h2 []gin.HandlerFunc
	for i := range h1 {
		h2 = append(h2, ToGinHandlerFunc(h1[i]))
	}
	return h2
}

func (g *RouterGroup) Use(middleware ...HandlerFunc) gin.IRoutes {
	return g.group.Use(handlerFuncToGinHandlerFunc(middleware...)...)
}

func (g *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group := g.group.Group(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
	return NewGroup(group)
}

func (g *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.Handle(httpMethod, relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.POST(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.GET(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.PUT(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.PATCH(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.DELETE(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.OPTIONS(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}

func (g *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return g.group.HEAD(relativePath, handlerFuncToGinHandlerFunc(handlers...)...)
}
