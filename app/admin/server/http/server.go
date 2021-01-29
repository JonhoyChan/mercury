package http

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/web"
	"mercury/app/admin/config"
	"mercury/app/admin/model"
	"mercury/app/admin/service"
	"mercury/x/ecode"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"net/http"
)

type httpServer struct {
	id  string
	e   *gin.Engine
	l   log.Logger
	srv *service.Service
}

func Init(c config.Provider, srv *service.Service) {
	microWeb := web.NewService(microx.InitWebOptions(c)...)
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
	v1 := ginx.NewGroup(s.e.Group("/admin/v1/"))
	{
		v1.GET("/clients", s.getClients)
		v1.GET("/clients/:id", s.getClient)
		v1.POST("/clients", s.createClient)
		v1.PUT("/clients", s.updateClient)
		v1.DELETE("/clients/:id", s.deleteClient)
	}
}

func (s *httpServer) getClients(c *ginx.Context) {
	c.Success(nil)
}

func (s *httpServer) getClient(c *ginx.Context) {
	id := c.Param("id")
	resp, err := s.srv.GetClient(c, id)
	if err != nil {
		s.l.Error("[getClient] failed to get client", "client_id", id, "error", err)
		c.Error(err)
		return
	}

	c.Success(resp)
}

func (s *httpServer) createClient(c *ginx.Context) {
	var req model.CreateClientReq
	if err := c.BindRequest(&req); err != nil {
		s.l.Error("[createClient] failed to bind request", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}

	resp, err := s.srv.CreateClient(c, req.FillToProto())
	if err != nil {
		s.l.Error("[createClient] failed to create client", "error", err)
		c.Error(err)
		return
	}

	c.Success(resp)
}

func (s *httpServer) updateClient(c *ginx.Context) {
	var req model.UpdateClientReq
	if err := c.BindRequest(&req); err != nil {
		s.l.Error("[updateClient] failed to bind request", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}

	if err := s.srv.UpdateClient(c, req.ID, req.FillToProto()); err != nil {
		s.l.Error("[createClient] failed to update client", "error", err)
		c.Error(err)
		return
	}

	c.Success(nil)
}

func (s *httpServer) deleteClient(c *ginx.Context) {
	id := c.Param("id")
	if err := s.srv.DeleteClient(c, id); err != nil {
		s.l.Error("[getClient] failed to delete client", "client_id", id, "error", err)
		c.Error(err)
		return
	}

	c.Success(nil)
}
