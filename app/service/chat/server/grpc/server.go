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
		srv.AuthenticateClientToken,
	))
	microServer := micro.NewService(opts...)
	microServer.Init()

	s := &grpcServer{
		s: srv,
	}

	if err := api.RegisterChatAdminHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
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

func (s *grpcServer) GenerateToken(ctx context.Context, req *api.GenerateTokenReq, resp *api.TokenResp) error {
	token, lifetime, err := s.s.GenerateToken(ctx, req)
	if err != nil {
		return err
	}

	resp.Token = token
	resp.Lifetime = lifetime
	return nil
}

func (s *grpcServer) CreateClient(ctx context.Context, req *api.CreateClientReq, resp *api.CreateClientResp) error {
	clientID, clientSecret, err := s.s.CreateClient(ctx, req)
	if err != nil {
		return err
	}

	resp.ClientID = clientID
	resp.ClientSecret = clientSecret
	return nil
}

func (s *grpcServer) UpdateClient(ctx context.Context, req *api.UpdateClientReq, resp *api.Empty) error {
	if err := s.s.UpdateClient(ctx, req); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) DeleteClient(ctx context.Context, req *api.DeleteClientReq, resp *api.Empty) error {
	if err := s.s.DeleteClient(ctx); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) CreateUser(ctx context.Context, req *api.CreateUserReq, resp *api.CreateUserResp) error {
	uid, err := s.s.CreateUser(ctx, req)
	if err != nil {
		return err
	}

	resp.UID = uid
	return nil
}

func (s *grpcServer) UpdateActivated(ctx context.Context, req *api.UpdateActivatedReq, resp *api.Empty) error {
	err := s.s.UpdateActivated(s.s.SetContextUser(ctx, req.UID), req.Activated)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) DeleteUser(ctx context.Context, req *api.DeleteUserReq, resp *api.Empty) error {
	if err := s.s.DeleteUser(s.s.SetContextUser(ctx, req.UID)); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) GenerateUserToken(ctx context.Context, req *api.GenerateUserTokenReq, resp *api.TokenResp) error {
	token, lifetime, err := s.s.GenerateUserToken(s.s.SetContextUser(ctx, req.UID))
	if err != nil {
		return err
	}

	resp.Token = token
	resp.Lifetime = lifetime
	return nil
}

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
