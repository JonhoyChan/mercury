package config

import (
	"mercury/x/config"
)

type Provider interface {
	config.DefaultProvider
	ConfigPath() string
	RepoPath() string
}
