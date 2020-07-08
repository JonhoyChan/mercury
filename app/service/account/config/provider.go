package config

import (
	"outgoing/x/config"
)

type Provider interface {
	config.DefaultProvider
	config.RegistryProvider
	config.DatabaseProvider
	config.RedisProvider
	config.HasherProvider
	config.GeneratorProvider
}
