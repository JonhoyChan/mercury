package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/server/grpc"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"mercury/app/comet/api"
	"mercury/app/comet/service"
	"mercury/app/comet/stats"
	"mercury/config"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/ginx"
	"mercury/x/log"
	"mercury/x/microx"
	"mercury/x/types"
	"mercury/x/websocket"
	"net/http"
	"strings"
)

type CometServer struct {
	id     string
	inst   *Instance
	log    log.Logger
	engine *gin.Engine
	srv    service.Servicer
}

func NewCometServer(inst *Instance, l log.Logger) *CometServer {
	return &CometServer{
		id:     uuid.New().String(),
		inst:   inst,
		log:    l,
		engine: gin.New(),
	}
}

func (s *CometServer) Serve(ctx context.Context) error {
	cfg := s.inst.cfg
	var err error
	if s.srv, err = service.NewService(s.log.New("service", "mercury.comet")); err != nil {
		return err
	}

	srvCfg, founded := cfg.GetService("mercury.comet")
	if !founded {
		return ecode.NewError("can not found \"mercury.job\" service config")
	}

	go s.RegisterRPC(cfg, srvCfg)

	opts := microx.DefaultWebOptions(srvCfg)
	opts = append(opts, web.Id(s.id))
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
	microWeb.Handle("/debug/vars", stats.Handler)

	return microWeb.Run()
}

func (s *CometServer) registerRouter() {
	registerMiddleware(s.engine)

	v1 := ginx.NewGroup(s.engine.Group("/chat/v1/"))
	{
		v1.GET("/channels", s.serveWebSocket)
	}
}

func (s *CometServer) serveWebSocket(c *ginx.Context) {
	conn, err := websocket.Upgrade(c.Writer, c.Request)
	if err != nil {
		c.Error(err)
		return
	}

	err = s.srv.SessionStore().NewSession(c, conn, s.id, s.srv)
	if err != nil {
		c.Error(err)
		return
	}
}

func (s *CometServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

func (s *CometServer) RegisterRPC(cfg *config.Config, srvCfg *config.Service) {
	opts := microx.DefaultServerOptions(srvCfg)
	opts = append(opts, server.Id(s.id), server.Address(srvCfg.RpcAddress()), server.WrapHandler(ecode.MicroHandlerFunc))

	if cfg.Registry.ETCD.Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range cfg.Registry.ETCD.Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, server.Registry(r))
	}

	microServer := grpc.NewServer(opts...)
	if err := microServer.Init(); err != nil {
		panic("unable to initialize server:" + err.Error())
	}

	if err := api.RegisterChatHandler(microServer, s); err != nil {
		panic("unable to register grpc server:" + err.Error())
	}

	if err := microServer.Start(); err != nil {
		panic("unable to start server:" + err.Error())
	}
}

func (s *CometServer) PushMessage(ctx context.Context, req *api.PushMessageReq, resp *api.Empty) error {
	//log.Info("[PushMessage] request is received")

	for _, sid := range req.SIDs {
		session := s.srv.SessionStore().Get(sid)
		if session != nil {
			go session.QueueOut(types.Operation(req.Operation), req.Data)
		}
	}
	return nil
}

func (s *CometServer) BroadcastMessage(ctx context.Context, req *api.BroadcastMessageReq, resp *api.Empty) error {
	//s.log.Info("[BroadcastMessage] request is received")

	sessions := s.srv.SessionStore().GetAll()
	for _, s := range sessions {
		if s != nil {
			go s.QueueOut(types.OperationBroadcast, req.Data)
		}
	}
	return nil
}
