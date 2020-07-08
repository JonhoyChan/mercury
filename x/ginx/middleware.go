package ginx

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http/httputil"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/stat"
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
	method := c.Request.Method

	// Process request
	c.Next()

	// Stop timer
	end := time.Now()
	latency := end.Sub(start)
	statusCode := c.Writer.Status()
	ecode := c.GetInt(ContextEcodeKey)
	clientIP := c.ClientIP()

	logrus.Infof("Method=%s Path=%s Code=%d IP=%s Time=%s Ecode=%d", method, path, statusCode, clientIP, latency.String(), ecode)
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
			pnc := x.Sprintf("[Recovery] %s panic recovered:\n%s\n%s\n%s", time.Now().Format("2006-01-02 15:04:05"), string(req), err, buf)
			logrus.Error(pnc)
			c.AbortWithStatus(500)
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
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Length", "Content-Type", "AppKey", "AccessToken"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
