package config

import (
	"fmt"
	"time"
)

const (
	defaultVersion          = "latest"
	defaultRegisterTTL      = "30s"
	defaultRegisterInterval = "15s"
	defaultHost             = "0.0.0.0"
)

type ServiceConfig map[string]interface{}

type Service struct {
	Name   string        `json:"name"`
	Config ServiceConfig `json:"config"`
}

func (s Service) ServiceName() string {
	return s.Name
}

func (s Service) Version() string {
	v, ok := s.Config["version"]
	if ok {
		return v.(string)
	}
	return defaultVersion
}

func (s Service) RegisterTTL() time.Duration {
	str := defaultRegisterTTL
	v, ok := s.Config["register_ttl"]
	if ok {
		str = v.(string)
	}
	d, _ := time.ParseDuration(str)
	return d
}

func (s Service) RegisterInterval() time.Duration {
	str := defaultRegisterInterval
	v, ok := s.Config["register_interval"]
	if ok {
		str = v.(string)
	}
	d, _ := time.ParseDuration(str)
	return d
}

func (s Service) Host() string {
	v, ok := s.Config["host"]
	if ok {
		return v.(string)
	}
	return defaultHost
}

func (s Service) Port() int {
	v, ok := s.Config["port"]
	if ok {
		return int(v.(float64))
	}
	return 0
}

func (s Service) RpcPort() int {
	v, ok := s.Config["rpc_port"]
	if ok {
		return int(v.(float64))
	}
	return 0
}

func (s Service) Address() string {
	return fmt.Sprintf("%s:%d", s.Host(), s.Port())
}

func (s Service) RpcAddress() string {
	return fmt.Sprintf("%s:%d", s.Host(), s.RpcPort())
}

func DefaultServices() []*Service {
	return []*Service{
		{
			Name: "mercury.infra",
			Config: map[string]interface{}{
				"version":           defaultVersion,
				"register_ttl":      defaultRegisterTTL,
				"register_interval": defaultRegisterInterval,
				"host":              defaultHost,
				"port":              9600,
			},
		},
		{
			Name: "mercury.admin",
			Config: map[string]interface{}{
				"version":           defaultVersion,
				"register_ttl":      defaultRegisterTTL,
				"register_interval": defaultRegisterInterval,
				"host":              defaultHost,
				"port":              9000,
			},
		},
		{
			Name: "mercury.comet",
			Config: map[string]interface{}{
				"version":           defaultVersion,
				"register_ttl":      defaultRegisterTTL,
				"register_interval": defaultRegisterInterval,
				"host":              defaultHost,
				"port":              9001,
				"rpc_port":          9002,
			},
		},
		{
			Name: "mercury.logic",
			Config: map[string]interface{}{
				"version":           defaultVersion,
				"register_ttl":      defaultRegisterTTL,
				"register_interval": defaultRegisterInterval,
				"host":              defaultHost,
				"port":              9011,
			},
		},
		{
			Name: "mercury.job",
			Config: map[string]interface{}{
				"version":           defaultVersion,
				"register_ttl":      defaultRegisterTTL,
				"register_interval": defaultRegisterInterval,
				"host":              defaultHost,
				"port":              9111,
			},
		},
	}
}
