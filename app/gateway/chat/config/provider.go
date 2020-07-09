package config

import (
	"outgoing/x/config"
)

type Provider interface {
	RPCAddress() string
	config.DefaultProvider
	config.RegistryProvider
}
