package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"mercury/app/admin/model"
	"mercury/app/admin/service"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"net/http"
	"strings"
)

type AdminServer struct {
	inst   *Instance
	log    log.Logger
	engine *gin.Engine
	srv    service.Servicer
}

func NewAdminServer(inst *Instance, l log.Logger) *AdminServer {
	return &AdminServer{
		inst:   inst,
		log:    l,
		engine: gin.New(),
	}
}

func (s *AdminServer) Serve(ctx context.Context) error {
	cfg := s.inst.cfg
	var err error
	if s.srv, err = service.NewService(s.log.New("service", "mercury.admin")); err != nil {
		return err
	}

	srvCfg, founded := cfg.GetService("mercury.admin")
	if !founded {
		return ecode.NewError("can not found \"mercury.admin\" service config")
	}

	opts := microx.DefaultWebOptions(srvCfg)
	if cfg.Registry.ETCD.Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range cfg.Registry.ETCD.Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, web.Registry(r))
	}

	microWeb := web.NewService(opts...)
	if err = microWeb.Init(); err != nil {
		return err
	}

	s.registerRouter()
	microWeb.Handle("/", s)

	return microWeb.Run()
}

func (s *AdminServer) registerRouter() {
	registerMiddleware(s.engine)

	v1 := ginx.NewGroup(s.engine.Group("/admin/v1/"))
	{
		v1.GET("/clients", s.getClients)
		v1.GET("/clients/:id", s.getClient)
		v1.POST("/clients", s.createClient)
		v1.PUT("/clients", s.updateClient)
		v1.DELETE("/clients/:id", s.deleteClient)
	}
}

func (s *AdminServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

func (s *AdminServer) getClients(c *ginx.Context) {
	// TODO
	c.Success(nil)
}

func (s *AdminServer) getClient(c *ginx.Context) {
	id := c.Param("id")
	resp, err := s.srv.GetClient(c, id)
	if err != nil {
		s.log.Error("[getClient] failed to get client", "client_id", id, "error", err)
		c.Error(err)
		return
	}

	c.Success(resp)
}

func (s *AdminServer) createClient(c *ginx.Context) {
	var req model.CreateClientReq
	if err := c.BindRequest(&req); err != nil {
		s.log.Error("[createClient] failed to bind request", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}

	resp, err := s.srv.CreateClient(c, req.FillToProto())
	if err != nil {
		s.log.Error("[createClient] failed to create client", "error", err)
		c.Error(err)
		return
	}

	c.Success(resp)
}

func (s *AdminServer) updateClient(c *ginx.Context) {
	var req model.UpdateClientReq
	if err := c.BindRequest(&req); err != nil {
		s.log.Error("[updateClient] failed to bind request", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}

	if err := s.srv.UpdateClient(c, req.ID, req.FillToProto()); err != nil {
		s.log.Error("[updateClient] failed to update client", "error", err)
		c.Error(err)
		return
	}

	c.Success(nil)
}

func (s *AdminServer) deleteClient(c *ginx.Context) {
	id := c.Param("id")
	if err := s.srv.DeleteClient(c, id); err != nil {
		s.log.Error("[deleteClient] failed to delete client", "client_id", id, "error", err)
		c.Error(err)
		return
	}

	c.Success(nil)
}
