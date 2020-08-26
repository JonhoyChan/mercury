package grpc

import (
	"context"
	"outgoing/app/gateway/chat/api"
	"outgoing/app/gateway/chat/config"
	"outgoing/app/gateway/chat/service"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"strings"

	"github.com/micro/go-micro/v2"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
)

type grpcServer struct {
	l   log.Logger
	srv *service.Service
}

// 注册服务
func Init(c config.Provider) {
	opts := []micro.Option{
		micro.Server(server.NewServer(server.Id(c.ID()))),
		micro.Name(c.Name()),
		micro.Version(c.Version()),
		micro.RegisterTTL(c.RegisterTTL()),
		micro.RegisterInterval(c.RegisterInterval()),
		micro.Address(c.RPCAddress()),
	}

	if c.Etcd().Enable {
		etcdv3Registry := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, micro.Registry(etcdv3Registry))
	}

	opts = append(opts, micro.WrapHandler(
		ratelimit.NewHandlerWrapper(1024),
		ecode.MicroHandlerFunc,
	))
	microServer := micro.NewService(opts...)
	microServer.Init()

	s := &grpcServer{}

	if err := api.RegisterChatHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	go func() {
		if err := microServer.Run(); err != nil {
			panic("unable to start service:" + err.Error())
		}
	}()
}

func (s *grpcServer) PublishMessage(ctx context.Context, req *api.Empty, resp *api.Empty) error {
	log.Info("[PublishMessage] request is received")

	session := s.srv.SessionStore.Get("req.SID")
	if session != nil {
	}
	return nil
}
