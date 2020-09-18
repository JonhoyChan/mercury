package http

import (
	"net/http"
	"outgoing/app/comet/stats"
	"outgoing/app/infra/config"
	"outgoing/app/infra/model"
	"outgoing/app/infra/service"
	"outgoing/x/ginx"
	"outgoing/x/log"

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
	opts := []web.Option{
		web.Name(c.Name()),
		web.Version(c.Version()),
		web.RegisterTTL(c.RegisterTTL()),
		web.RegisterInterval(c.RegisterInterval()),
		web.Address(c.Address()),
	}

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
	v1 := ginx.NewGroup(s.e.Group("/infra/v1/"))
	{
		v1.GET("/config", s.loadConfig)
	}
}

func (s *httpServer) loadConfig(c *ginx.Context) {
	name := c.Query("name")

	ciphertext, err := s.srv.LoadConfig(name)
	if err != nil {
		c.Error(err)
		return
	}

	resp := &model.Config{Ciphertext: ciphertext}
	c.Success(resp)
}
