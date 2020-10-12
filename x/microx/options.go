package microx

import (
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/broker/stan/v2"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"mercury/x"
	"mercury/x/config"
	"strings"
)

func InitDefaultServerOptions(c config.DefaultProvider) []server.Option {
	opts := []server.Option{
		server.Name(c.Name()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.Address()),
	}
	return opts
}

func InitDefaultWebOptions(c config.DefaultProvider) []web.Option {
	opts := []web.Option{
		web.Name(c.Name()),
		web.Version(c.Version()),
		web.RegisterTTL(c.RegisterTTL()),
		web.RegisterInterval(c.RegisterInterval()),
		web.Address(c.Address()),
	}
	return opts
}

type configProvider interface {
	config.DefaultProvider
	config.RegistryProvider
	config.BrokerProvider
}

func InitServerOptions(c configProvider) []server.Option {
	opts := []server.Option{
		server.Name(c.Name()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.Address()),
	}

	// 判断是否使用了etcd作为服务注册
	if c.Etcd().Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, server.Registry(r))
	}

	if c.Stan().Enable {
		// 创建一个新stanBroker实例
		b := stan.NewBroker(
			// 设置stan集群的地址
			broker.Addrs(c.Stan().Addresses...),
			stan.ConnectRetry(true),
			// 设置stan集群标识
			stan.ClusterID(c.Stan().ClusterID),
			// 设置订阅者使用的永久名
			stan.DurableName(c.Stan().DurableName),
		)

		if err := b.Init(); err != nil {
			panic("unable to init stan broker:" + err.Error())
		}

		if err := b.Connect(); err != nil {
			panic("unable to connect to stan broker:" + err.Error())
		}

		opts = append(opts, server.Broker(b))
	}

	return opts
}

func InitMicroOptions(c configProvider) []micro.Option {
	opts := []micro.Option{
		micro.Name(c.Name()),
		micro.Version(c.Version()),
		micro.RegisterTTL(c.RegisterTTL()),
		micro.RegisterInterval(c.RegisterInterval()),
		micro.Address(c.Address()),
	}

	// 判断是否使用了etcd作为服务注册
	if c.Etcd().Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, micro.Registry(r))
	}

	if c.Stan().Enable {
		// 创建一个新stanBroker实例
		b := stan.NewBroker(
			// 设置stan集群的地址
			broker.Addrs(c.Stan().Addresses...),
			stan.ConnectRetry(true),
			// 设置stan集群标识
			stan.ClusterID(c.Stan().ClusterID),
			// 设置订阅者使用的永久名
			stan.DurableName(c.Stan().DurableName),
		)

		if err := b.Init(); err != nil {
			panic("unable to init stan broker:" + err.Error())
		}

		if err := b.Connect(); err != nil {
			panic("unable to connect to stan broker:" + err.Error())
		}

		opts = append(opts, micro.Broker(b))
		broker.DefaultBroker = b
	}

	return opts
}

type configProviderWithoutBroker interface {
	config.DefaultProvider
	config.RegistryProvider
}

func InitServerOptionsWithoutBroker(c configProviderWithoutBroker) []server.Option {
	opts := []server.Option{
		server.Name(c.Name()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.Address()),
	}

	// 判断是否使用了etcd作为服务注册
	if c.Etcd().Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, server.Registry(r))
	}

	return opts
}

func InitWebOptionsWithoutBroker(c configProviderWithoutBroker) []web.Option {
	opts := []web.Option{
		web.Name(c.Name()),
		web.Version(c.Version()),
		web.RegisterTTL(c.RegisterTTL()),
		web.RegisterInterval(c.RegisterInterval()),
		web.Address(c.Address()),
	}

	// 判断是否使用了etcd作为服务注册
	if c.Etcd().Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, web.Registry(r))
	}

	return opts
}
