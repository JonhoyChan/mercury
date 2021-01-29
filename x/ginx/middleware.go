package ginx

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/stat"
	"net/http/httputil"
	"runtime"
	"time"
)

var (
	stats = stat.HTTPServer
)

func loggerHandler(c *gin.Context) {
	// Start timer
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}

	// Process request
	c.Next()

	// Stop timer
	end := time.Now()
	code := c.GetInt(ContextEcodeKey)
	if code == 0 {
		code = 200
	}

	log.Debug("[Logger] request processing completed",
		"method", c.Request.Method,
		"path", path,
		"status_code", c.Writer.Status(),
		"ip", c.ClientIP(),
		"time", end.Sub(start).String(),
		"ecode", code,
	)
}

func Logger() gin.HandlerFunc {
	return loggerHandler
}

func recoverHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			req, _ := httputil.DumpRequest(c.Request, false)
			log.Error("[Recovery] panic recovered", "request", req, "error", err, "buffer", buf)
			c.AbortWithStatus(StatusInternalServerError)
		}
	}()
	c.Next()
}

func Recovery() gin.HandlerFunc {
	return recoverHandler
}

func NoMethodHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := &Context{c}
		context.Error(ecode.ErrMethodNotAllowed)
	}
}

func NoRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := &Context{c}
		context.Error(ecode.ErrNotFound)
	}
}

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
