package grpc

import (
	"context"
	"outgoing/app/service/account/api"
	"outgoing/app/service/account/config"
	"outgoing/app/service/account/service"
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

	if err := api.RegisterAccountHandler(microServer, s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	// Run service
	go func() {
		if err := microServer.Start(); err != nil {
			panic("unable to start service:" + err.Error())
		}
	}()
}

// 创建新用户
func (s *grpcServer) Register(ctx context.Context, req *api.RegisterReq, resp *api.RegisterResp) error {
	m := strings.Split(req.Mobile, " ")
	if len(m) != 2 {
		return ecode.Wrap(ecode.ErrWrongParameter, "无效的参数")
	}

	mobile := m[1]
	if m[0] == "+86" {
		if match := x.MatchMobile(mobile); !match {
			return ecode.Wrap(ecode.ErrPhoneNumber, "无效的手机号码")
		}
	}

	if match := x.MatchIP(req.Ip); !match {
		return ecode.ErrIPAddress
	}

	uid, vid, err := s.s.Register(ctx, req.Mobile, req.Ip)
	if err != nil {
		return err
	}

	resp.UID = uid
	return nil
}

// 用户登录
func (s *grpcServer) Login(ctx context.Context, req *api.LoginReq, resp *api.LoginResp) error {
	if match := x.MatchIP(req.Ip); !match {
		return ecode.Wrap(ecode.ErrIPAddress, "无效的IP地址")
	}

	var (
		uid, vid string
		err      error
	)
	m := strings.Split(req.Input, " ")
	if len(m) == 2 {
		mobile := m[1]
		if m[0] == "+86" {
			if match := x.MatchMobile(mobile); !match {
				return ecode.Wrap(ecode.ErrPhoneNumber, "无效的手机号码")
			}
		}

		uid, vid, err = s.s.LoginViaMobile(ctx, mobile, req.Captcha, req.Password, req.Version, req.DeviceId, req.Ip)
	} else {

	}
	if err != nil {
		return err
	}

	resp.UID = uid
	return nil
}
