package microx

import (
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/web"
	"time"
)

type ConfigProvider interface {
	ServiceName() string
	Version() string
	RegisterTTL() time.Duration
	RegisterInterval() time.Duration
	Address() string
}

func DefaultWebOptions(c ConfigProvider) []web.Option {
	opts := []web.Option{
		web.Name(c.ServiceName()),
		web.Version(c.Version()),
		web.RegisterTTL(c.RegisterTTL()),
		web.RegisterInterval(c.RegisterInterval()),
		web.Address(c.Address()),
	}
	return opts
}

func DefaultServerOptions(c ConfigProvider) []server.Option {
	opts := []server.Option{
		server.Name(c.ServiceName()),
		server.Version(c.Version()),
		server.RegisterTTL(c.RegisterTTL()),
		server.RegisterInterval(c.RegisterInterval()),
		server.Address(c.Address()),
	}
	return opts
}

func DefaultMicroOptions(c ConfigProvider) []micro.Option {
	opts := []micro.Option{
		micro.Name(c.ServiceName()),
		micro.Version(c.Version()),
		micro.RegisterTTL(c.RegisterTTL()),
		micro.RegisterInterval(c.RegisterInterval()),
		micro.Address(c.Address()),
	}
	return opts
}
