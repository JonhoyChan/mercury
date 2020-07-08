package config

import (
	"outgoing/x/config"
)

type Provider interface {
	config.DefaultProvider
	config.RegistryProvider
	config.BrokerProvider
}
