package config

import (
	"outgoing/x/config"
)

type Provider interface {
	config.DefaultProvider
	ConfigPassphrase() string
	ConfigPath() string
}
