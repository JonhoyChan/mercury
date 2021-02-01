package lib

import (
	"context"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-plugins/broker/stan/v2"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"mercury/app/job/service"
	"mercury/config"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/microx"
	"strings"
)

type JobServer struct {
	inst *Instance
	log  log.Logger
	srv  service.Servicer
}

func NewJobServer(inst *Instance, l log.Logger) *JobServer {
	return &JobServer{
		inst: inst,
		log:  l,
	}
}

func (s *JobServer) Serve(ctx context.Context) error {
	cfg := config.NewProviderConfig(s.inst.cfg)
	var err error
	if s.srv, err = service.NewService(cfg, s.log.New("service", "mercury.job")); err != nil {
		return err
	}

	srvCfg, founded := cfg.GetService("mercury.job")
	if !founded {
		return ecode.NewError("can not found \"mercury.job\" service config")
	}

	opts := microx.DefaultServerOptions(srvCfg)
	// 判断是否使用了etcd作为服务注册
	if cfg.Registry().ETCD.Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range cfg.Registry().ETCD.Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, server.Registry(r))
	}

	if cfg.Broker().Stan.Enable {
		// 创建一个新stanBroker实例
		b := stan.NewBroker(
			// 设置stan集群的地址
			broker.Addrs(cfg.Broker().Stan.Addresses...),
			stan.ConnectRetry(true),
			// 设置stan集群标识
			stan.ClusterID(cfg.Broker().Stan.ClusterID),
			// 设置订阅者使用的永久名
			stan.DurableName(cfg.Broker().Stan.DurableName),
		)

		if err := b.Init(); err != nil {
			panic("unable to init stan broker:" + err.Error())
		}

		if err := b.Connect(); err != nil {
			panic("unable to connect to stan broker:" + err.Error())
		}

		broker.DefaultBroker = b
		opts = append(opts, server.Broker(b))
	}

	microServer := server.NewServer(opts...)
	if err := microServer.Init(); err != nil {
		return err
	}
	s.srv.Init(microServer.Options())

	return microServer.Start()
}
