package http

import (
	"mercury/app/comet/config"
	"mercury/app/comet/service"
	"mercury/app/comet/stats"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"mercury/x/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/web"
)

type httpServer struct {
	id  string
	e   *gin.Engine
	l   log.Logger
	srv *service.Service
}

func Init(c config.Provider, srv *service.Service) {
	opts := append(microx.InitWebOptionsWithoutBroker(c), web.Id(c.ID()))
	microWeb := web.NewService(opts...)

	// Initialize service
	if err := microWeb.Init(); err != nil {
		panic("unable to initialize service:" + err.Error())
	}

	s := &httpServer{
		id:  microWeb.Options().Id,
		e:   gin.New(),
		l:   c.Logger(),
		srv: srv,
	}
	s.middleware()
	s.setupRouter()

	microWeb.Handle("/", s)
	stats.Init(microWeb, "/debug/vars")

	// Run service
	go func() {
		if err := microWeb.Run(); err != nil {
			panic("unable to run http service:" + err.Error())
		}
	}()
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.e.ServeHTTP(w, req)
}

func (s *httpServer) middleware() {
	s.e.NoMethod(ginx.NoMethodHandler())
	s.e.NoRoute(ginx.NoRouteHandler())

	s.e.Use(ginx.Recovery(), ginx.Logger(), ginx.CORS())
}

func (s *httpServer) setupRouter() {
	v1 := ginx.NewGroup(s.e.Group("/chat/v1/"))
	{
		v1.GET("/channels", s.serveWebSocket)
	}
}

func (s *httpServer) serveWebSocket(c *ginx.Context) {
	conn, err := websocket.Upgrade(c.Writer, c.Request)
	if err != nil {
		c.Error(err)
		return
	}

	err = s.srv.SessionStore.NewSession(c, conn, s.id, s.srv)
	if err != nil {
		c.Error(err)
		return
	}
}
