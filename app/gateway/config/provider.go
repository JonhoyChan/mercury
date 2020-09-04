package config

import (
	"outgoing/x/config"
)

type Provider interface {
	ID() string
	RPCAddress() string
	config.DefaultProvider
	config.RegistryProvider
}
