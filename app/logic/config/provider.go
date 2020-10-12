package config

import (
	"mercury/x/config"
)

type Provider interface {
	config.DefaultProvider
	config.RegistryProvider
	config.BrokerProvider
	config.DatabaseProvider
	config.RedisProvider
	config.AuthenticatorProvider
	config.HasherProvider
	config.GeneratorProvider
	PushMessageTopic() string
}
