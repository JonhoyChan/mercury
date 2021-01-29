package config

import "time"

type Registry struct {
	ETCD RegistryETCD `json:"etcd"`
}

type RegistryETCD struct {
	Enable    bool          `json:"enable"`
	Addresses []string      `json:"addresses"`
	Timeout   time.Duration `json:"timeout"`
}

func DefaultRegistry() *Registry {
	return &Registry{}
}
