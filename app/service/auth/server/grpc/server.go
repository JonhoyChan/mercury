package grpc

import (
	"context"
	"outgoing/app/service/auth/api"
	"outgoing/app/service/auth/auth"
	"outgoing/app/service/auth/config"
	"outgoing/app/service/auth/service"
	"outgoing/x"
	"outgoing/x/ecode"
	"strings"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
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

	wrapHandlers := []server.HandlerWrapper{
		ecode.MicroHandlerFunc,
		ratelimit.NewHandlerWrapper(1024),
	}
	opts = append(opts, micro.WrapHandler(wrapHandlers...))

	microService := micro.NewService(opts...)
	microService.Init()

	s := &grpcServer{
		s: srv,
	}

	if err := api.RegisterAuthHandler(microService.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	// Run service
	go func() {
		if err := microService.Run(); err != nil {
			panic("unable to run grpc service:" + err.Error())
		}
	}()
}

func getHandlerType(handlerType api.HandlerType) (t auth.HandlerType) {
	switch handlerType {
	case api.HandlerType_HandlerTypeToken:
		t = auth.Token
	case api.HandlerType_HandlerTypeJWT:
		t = auth.JWT
	}

	return
}

// 根据 Record 生成一个新的 Token
func (s *grpcServer) GenerateToken(ctx context.Context, req *api.GenerateTokenReq, resp *api.GenerateTokenResp) error {
	if req.Record == nil || req.Record.Uid == "" {
		return ecode.ErrWrongParameter
	}

	t := getHandlerType(req.HandlerType)

	token, err := s.s.GenerateToken(ctx, t, req.Record)
	if err != nil {
		return err
	}

	resp.Token = token
	return nil
}

// 验证 Token 并返回 Record
func (s *grpcServer) Authenticate(ctx context.Context, req *api.AuthenticateReq, resp *api.AuthenticateResp) error {
	if req.Token == "" {
		return ecode.ErrWrongParameter
	}

	t := getHandlerType(req.HandlerType)

	record, err := s.s.Authenticate(ctx, t, req.Token)
	if err != nil {
		return err
	}

	resp.Record = record
	return nil
}
