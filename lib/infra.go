package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/web"
	"mercury/app/infra/service"
	"mercury/x/ecode"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"net/http"
)

type InfraServer struct {
	inst   *Instance
	log    log.Logger
	engine *gin.Engine
	srv    service.Servicer
}

func NewInfraServer(inst *Instance, l log.Logger) *InfraServer {
	return &InfraServer{
		inst:   inst,
		log:    l,
		engine: gin.New(),
	}
}

func (s *InfraServer) Serve(ctx context.Context) error {
	cfg := s.inst.cfg
	var err error
	if s.srv, err = service.NewService(cfg, s.log.New("service", "mercury.infra")); err != nil {
		return err
	}

	srvCfg, founded := cfg.GetService("mercury.infra")
	if !founded {
		return ecode.NewError("can not found \"mercury.infra\" service config")
	}
	microWeb := web.NewService(microx.DefaultWebOptions(srvCfg)...)
	if err = microWeb.Init(); err != nil {
		return err
	}

	s.registerRouter()
	microWeb.Handle("/", s)

	return microWeb.Run()
}

func (s *InfraServer) registerRouter() {
	registerMiddleware(s.engine)

	v1 := ginx.NewGroup(s.engine.Group("/infra/v1/"))
	{
		v1.GET("/config", s.loadConfig)
	}
}

func (s *InfraServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

func (s *InfraServer) loadConfig(c *ginx.Context) {
	resp, err := s.srv.LoadConfig()
	if err != nil {
		c.Error(err)
		return
	}

	c.Success(resp)
}
