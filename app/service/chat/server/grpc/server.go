package grpc

import (
	"context"
	"outgoing/app/service/chat/api"
	"outgoing/app/service/chat/config"
	"outgoing/app/service/chat/service"
	"outgoing/x"
	"outgoing/x/ecode"
	"strings"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
)

type grpcServer struct {
	s *service.Service
}

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	opts := []server.Option{
		server.Name(c.Name()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.Address()),
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

	s := &grpcServer{
		s: srv,
	}

	if err := api.RegisterChatHandler(microServer, s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	// Run service
	go func() {
		if err := microServer.Start(); err != nil {
			panic("unable to start service:" + err.Error())
		}
	}()
}

// Connect a connection
func (s *grpcServer) Connect(ctx context.Context, req *api.ConnectReq, resp *api.Empty) error {
	if req.UID == "" || req.SID == "" || req.ServerID == "" {
		return ecode.ErrWrongParameter
	}

	err := s.s.Connect(ctx, req.UID, req.SID, req.ServerID)
	if err != nil {
		return err
	}

	return nil
}

// Disconnect a connection
func (s *grpcServer) Disconnect(ctx context.Context, req *api.DisconnectReq, resp *api.Empty) error {
	if req.UID == "" || req.SID == "" {
		return ecode.ErrWrongParameter
	}

	err := s.s.Disconnect(ctx, req.UID, req.SID)
	if err != nil {
		return err
	}

	return nil
}

// Heartbeat a connection
func (s *grpcServer) Heartbeat(ctx context.Context, req *api.HeartbeatReq, resp *api.Empty) error {
	if req.UID == "" || req.SID == "" || req.ServerID == "" {
		return ecode.ErrWrongParameter
	}

	err := s.s.Heartbeat(ctx, req.UID, req.SID, req.ServerID)
	if err != nil {
		return err
	}

	return nil
}
