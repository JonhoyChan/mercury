package config

import (
	"outgoing/x/config"
)

type Provider interface {
	config.DefaultProvider
	config.RegistryProvider
	config.DatabaseProvider
	config.RedisProvider
	config.AuthenticatorProvider
	config.HasherProvider
	config.GeneratorProvider
}
