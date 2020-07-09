package grpc

import (
	"context"
	"outgoing/app/gateway/chat/api"
	"outgoing/app/gateway/chat/config"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"strings"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
)

type grpcServer struct{}

// 注册服务
func Init(c config.Provider) {
	opts := []server.Option{
		server.Id(c.ID()),
		server.Name(c.Name()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.RPCAddress()),
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
		opts = append(opts, server.Registry(etcdv3Registry))
	}

	wrapHandlers := []server.Option{
		server.WrapHandler(ecode.MicroHandlerFunc),
		server.WrapHandler(ratelimit.NewHandlerWrapper(1024)),
	}
	opts = append(opts, wrapHandlers...)

	microServer := server.NewServer(opts...)
	if err := microServer.Init(); err != nil {
		panic("unable to initialize service:" + err.Error())
	}

	s := &grpcServer{}

	if err := api.RegisterChatHandler(microServer, s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	// Run service
	go func() {
		if err := microServer.Start(); err != nil {
			panic("unable to run grpc service:" + err.Error())
		}
	}()
}

func (s *grpcServer) PublishMessage(ctx context.Context, req *api.Empty, resp *api.Empty) error {
	log.Info("[PublishMessage] request is received")
	return nil
}
