package ginx

import "github.com/gin-gonic/gin"

type HandlerFunc func(c *Context)

func ToGinHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := new(Context)
		context.Context = c

		handler(context)
	}
}
