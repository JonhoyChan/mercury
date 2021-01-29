package lib

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/web"
	"io/ioutil"
	"mercury/app/infra/service"
	"mercury/x/ecode"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"net/http"
	"time"
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
		v1.POST("/file", s.addFile)
		v1.GET("/file/:hash", s.catFile)
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

func (s *InfraServer) addFile(c *ginx.Context) {
	file, err := c.FormFile("mercury-file")
	if err != nil {
		//log.Error("[uploadFile] failed to bind form file", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}

	f, err := file.Open()
	if err != nil {
		//log.Error("[uploadFile] failed to open file", "error", err)
		c.Error(ecode.ErrBadRequest)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		//log.Error("[uploadFile] failed to read file data", "error", err)
		c.Error(ecode.ErrBadRequest)
	}
	_ = f.Close()

	resp, err := s.srv.AddFile(data)
	if err != nil {
		c.Error(err)
		return
	}

	c.Success(resp)
}

func (s *InfraServer) catFile(c *ginx.Context) {
	hash := c.Param("hash")
	data, err := s.srv.CatFile(hash)
	if err != nil {
		//log.Error("[catFile] failed to cat file", "hash", hash, "error", err)
		c.Error(err)
	}

	c.Writer.Header().Set("Cache-Control", "public, max-age=29030400, immutable")
	http.ServeContent(c.Writer, c.Request, hash, time.Now(), bytes.NewReader(data))
}
