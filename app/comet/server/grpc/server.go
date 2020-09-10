package grpc

import (
	"context"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/server/grpc"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"outgoing/app/gateway/api"
	"outgoing/app/gateway/config"
	"outgoing/app/gateway/service"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"outgoing/x/types"
	"strings"
)

type grpcServer struct {
	l   log.Logger
	srv *service.Service
}

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	opts := []server.Option{
		server.Id(c.ID()),
		server.Name(c.Name()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.RPCAddress()),
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
		opts = append(opts, server.Registry(etcdv3Registry))
	}

	opts = append(opts, server.WrapHandler(ecode.MicroHandlerFunc))
	microServer := grpc.NewServer(opts...)
	if err := microServer.Init(); err != nil {
		panic("unable to initialize server:" + err.Error())
	}

	s := &grpcServer{
		srv: srv,
	}

	if err := api.RegisterChatHandler(microServer, s); err != nil {
		panic("unable to register grpc server:" + err.Error())
	}

	go func() {
		if err := microServer.Start(); err != nil {
			panic("unable to start server:" + err.Error())
		}
	}()
}

func (s *grpcServer) PushMessage(ctx context.Context, req *api.PushMessageReq, resp *api.Empty) error {
	log.Info("[PushMessage] request is received")

	for _, sid := range req.SIDs {
		session := s.srv.SessionStore.Get(sid)
		if session != nil {
			go session.QueueOut(types.OperationPush, req.Data)
		}
	}
	return nil
}

func (s *grpcServer) BroadcastMessage(ctx context.Context, req *api.BroadcastMessageReq, resp *api.Empty) error {
	log.Info("[BroadcastMessage] request is received")

	sessions := s.srv.SessionStore.GetAll()
	for _, s := range sessions {
		if s != nil {
			go s.QueueOut(types.OperationBroadcast, req.Data)
		}
	}
	return nil
}
