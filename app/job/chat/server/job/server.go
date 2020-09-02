package job

import (
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-plugins/broker/stan/v2"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"outgoing/app/job/chat/config"
	"outgoing/app/job/chat/service"
	"outgoing/x"
	"strings"
)

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
			panic("unable to init stan broker:" + err.Error())
		}

		if err := stanBroker.Connect(); err != nil {
			panic("unable to connect to stan broker:" + err.Error())
		}

		opts = append(opts, server.Broker(stanBroker))
	}

	microServer := server.NewServer(opts...)
	if err := microServer.Init(); err != nil {
		panic("unable to initialize server:" + err.Error())
	}

	srv.WithRegistry(microServer.Options().Registry)
	srv.WithBroker(microServer.Options().Broker)
	srv.WatchComet()

	// Run service
	go func() {
		if err := microServer.Start(); err != nil {
			panic("unable to start service:" + err.Error())
		}
	}()
}
