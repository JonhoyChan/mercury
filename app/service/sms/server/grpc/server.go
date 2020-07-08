package grpc

import (
	"context"
	"strings"
	"outgoing/app/service/main/sms/api"
	"outgoing/app/service/main/sms/config"
	"outgoing/app/service/main/sms/service"
	"outgoing/x"
	"outgoing/x/ecode"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/broker/stan/v2"
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

	if c.Stan().Enable {
		// 创建一个新stanBroker实例
		stanBroker := stan.NewBroker(
			// 设置stan集群的地址
			broker.Addrs(c.Stan().Addresses...),
			stan.ConnectRetry(true),
			// 设置stan集群标识
			stan.ClusterID(c.Stan().ClusterID),
			// 设置订阅者使用的永久名
			stan.DurableName(c.Stan().DurableName),
		)

		if err := stanBroker.Init(); err != nil {
			panic("unable to initialize the broker for stan:" + err.Error())
		}

		if err := stanBroker.Connect(); err != nil {
			panic("unable to connect the broker for stan:" + err.Error())
		}

		opts = append(opts, micro.Broker(stanBroker))

		// 为了能使用broker包提供的操作接口，设置全局变量broker.DefaultBroker指向这个stan插件实例。
		broker.DefaultBroker = stanBroker
	}

	microService := micro.NewService(opts...)
	microService.Init(micro.WrapHandler(ratelimit.NewHandlerWrapper(1024)))

	s := &grpcServer{
		s: srv,
	}

	if err := api.RegisterSMSHandler(microService.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	// Run service
	go func() {
		if err := microService.Run(); err != nil {
			panic("unable to run grpc service:" + err.Error())
		}
	}()
}

// 发送短信
func (s *grpcServer) SendCaptcha(ctx context.Context, req *api.SendReq, resp *api.SendResp) error {
	m := strings.Split(req.Mobile, " ")
	if len(m) != 2 {
		return ecode.ErrWrongParameter
	}

	country := m[0]
	if strings.HasPrefix(country, "+") {
		country = strings.TrimPrefix(country, "+")
	}
	mobile := m[1]
	if country == "86" {
		if match := x.MatchMobile(mobile); !match {
			return ecode.ErrPhoneNumber
		}
	}

	err := s.s.Send(ctx, req.Uid, country, mobile)
	if err != nil {
		return err
	}
	return nil
}
