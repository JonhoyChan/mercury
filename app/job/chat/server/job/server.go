package job

import (
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"outgoing/app/job/chat/config"
	"outgoing/app/job/chat/service"
	"outgoing/x"
	"strings"
)

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

	microService := micro.NewService(opts...)
	microService.Init()

	srv.WithRegistry(microService.Options().Registry)
	srv.WatchComet()

	// Run service
	go func() {
		if err := microService.Run(); err != nil {
			panic("unable to run grpc service:" + err.Error())
		}
	}()
}

func Close() {

}
