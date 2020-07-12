package http

import (
	"net/http"
	"outgoing/app/gateway/account/config"
	"outgoing/app/gateway/account/model"
	"outgoing/app/gateway/account/service"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/ginx"
	"outgoing/x/log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
)

type httpServer struct {
	e *gin.Engine
	s *service.Service
	l log.Logger
}

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	opts := []web.Option{
		web.Name(c.Name()),
		web.Version(c.Version()),
		web.RegisterTTL(c.RegisterTTL()),
		web.RegisterInterval(c.RegisterInterval()),
		web.Address(c.Address()),
	}

	// 判断是否使用了etcd作为服务注册
	if c.Etcd().Enable {
		etcdv3Registry := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, web.Registry(etcdv3Registry))
	}

	microWeb := web.NewService(opts...)

	// Initialize service
	if err := microWeb.Init(); err != nil {
		panic("unable to  initialize service:" + err.Error())
	}

	s := &httpServer{
		e: gin.New(),
		s: srv,
		l: c.Logger(),
	}
	s.middleware()
	s.setupRouter()

	microWeb.Handle("/", s)

	// Run service
	if err := microWeb.Run(); err != nil {
		panic("unable to run http service:" + err.Error())
	}
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.e.ServeHTTP(w, req)
}

func (s *httpServer) middleware() {
	s.e.Use(ginx.Recovery(), ginx.Logger(), ginx.CORS())
}

func (s *httpServer) setupRouter() {
	v1 := ginx.NewGroup(s.e.Group("/api/v1/"))
	{
		// 创建新用户（注册）
		v1.POST("/users", s.register)
		// 更新用户信息
		v1.PUT("/users")
		// 更新用户部分信息
		v1.PATCH("/users")
		// 删除用户（注销）
		v1.DELETE("/users")

		// 获取会话信息
		v1.GET("/sessions")
		// 创建新的会话（登录）
		v1.POST("/sessions", s.login)
		// 销毁当前会话（登出）
		v1.DELETE("/sessions")
	}
}

func (s *httpServer) register(c *ginx.Context) {
	var req model.RegisterReq
	if err := c.BindRequest(&req); err != nil {
		s.l.Error("register error", "error", err)
		c.Error(ecode.ErrBadRequest)
	}

	resp, err := s.s.Register(c, req.Mobile, c.ClientIP())
	if err != nil {
		s.l.Error("register error", "error", err)
		c.Error(err)
		return
	}

	c.Success(resp)
}

func (s *httpServer) login(c *ginx.Context) {
	var req model.LoginReq
	if err := c.BindRequest(&req); err != nil {
		s.l.Error("login error", "error", err)
		c.Error(ecode.ErrBadRequest)
		return
	}

	resp, err := s.s.Login(c, req.Input, req.Captcha, req.Password, c.ClientIP())
	if err != nil {
		s.l.Error("login error", "error", err)
		c.Error(err)
		return
	}

	c.Success(resp)
}
