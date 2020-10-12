package http

import (
	"bytes"
	"io/ioutil"
	"mercury/app/infra/config"
	"mercury/app/infra/model"
	"mercury/app/infra/service"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"net/http"
	"time"

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
	microWeb := web.NewService(microx.InitDefaultWebOptions(c)...)
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
	v1 := ginx.NewGroup(s.e.Group("/infra/v1/"))
	{
		v1.GET("/config", s.loadConfig)
		v1.POST("/file", s.uploadFile)
		v1.GET("/file/:hash", s.getFile)
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

func (s *httpServer) uploadFile(c *ginx.Context) {
	file, err := c.FormFile("mercury-file")
	if err != nil {
		s.l.Error("[uploadFile] failed to bind form file", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}
	f, err := file.Open()
	if err != nil {
		s.l.Error("[uploadFile] failed to open file", "error", err)
		c.Error(ecode.ErrBadRequest)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		s.l.Error("[uploadFile] failed to read file data", "error", err)
		c.Error(ecode.ErrBadRequest)
	}
	_ = f.Close()

	ft := x.GetFileType(data)
	switch ft {
	case "jpg", "png", "gif":
		resp, err := s.srv.AddImages(data, file.Filename)
		if err != nil {
			s.l.Error("[uploadFile] failed to add image", "error", err)
			c.Error(err)
			return
		}
		c.Success(resp)
	case "mp4": // FIXME
		resp, err := s.srv.AddVideos(data, file.Filename)
		if err != nil {
			s.l.Error("[uploadFile] failed to add video", "error", err)
			c.Error(err)
			return
		}
		c.Success(resp)
	default:
		// FIXME
		resp, err := s.srv.AddVideos(data, file.Filename)
		if err != nil {
			s.l.Error("[uploadFile] failed to add video", "error", err)
			c.Error(err)
			return
		}
		c.Success(resp)
		//c.Error(ecode.ErrBadRequest.ResetMessage(x.Sprintf("unsupported '%s' file type", ft)))
	}
}

func (s *httpServer) getFile(c *ginx.Context) {
	hash := c.Param("hash")

	data, err := s.srv.CatFile(hash)
	if err != nil {
		s.l.Error("[getFile] failed to get file", "hash", hash, "error", err)
		c.Error(err)
		return
	}

	c.Writer.Header().Set("Cache-Control", "public, max-age=29030400, immutable")
	//c.Writer.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(c.Writer, c.Request, hash, time.Now(), bytes.NewReader(data))
}
