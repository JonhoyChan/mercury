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

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
)

type grpcServer struct {
	s *service.Service
}

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	opts := []micro.Option{
		micro.Name(c.Name()),
		micro.Version(c.Version()),
		micro.RegisterTTL(c.RegisterTTL()),
		micro.RegisterInterval(c.RegisterInterval()),
		micro.Address(c.Address()),
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
		opts = append(opts, micro.Registry(etcdv3Registry))
	}

	opts = append(opts, micro.WrapHandler(
		ratelimit.NewHandlerWrapper(1024),
		ecode.MicroHandlerFunc,
	))
	microServer := micro.NewService(opts...)
	microServer.Init()

	s := &grpcServer{
		s: srv,
	}

	if err := api.RegisterChatHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	go func() {
		if err := microServer.Run(); err != nil {
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
