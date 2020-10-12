package grpc

import (
	"context"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/server/grpc"
	"mercury/app/comet/api"
	"mercury/app/comet/config"
	"mercury/app/comet/service"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/microx"
	"mercury/x/types"
)

type grpcServer struct {
	l   log.Logger
	srv *service.Service
}

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	opts := append(microx.InitServerOptionsWithoutBroker(c), server.Id(c.ID()), server.WrapHandler(ecode.MicroHandlerFunc))
	microServer := grpc.NewServer(opts...)
	if err := microServer.Init(server.Address(c.RPCAddress())); err != nil {
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
			go session.QueueOut(types.Operation(req.Operation), req.Data)
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
