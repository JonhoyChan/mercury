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
}

type RegistryProvider interface {
	Etcd() *config.EtcdConfig
}

type BrokerProvider interface {
	Stan() *config.StanConfig
}

type DatabaseProvider interface {
	Database() *config.DatabaseConfig
}

type RedisProvider interface {
	Redis() *config.RedisConfig
}

type HasherProvider interface {
	HasherArgon2() *config.HasherArgon2Config
	HasherBCrypt() *config.HasherBCryptConfig
}

type GeneratorProvider interface {
	GeneratorUid() *config.GeneratorUidConfig
}

type AuthenticatorProvider interface {
	AuthenticatorToken() *config.AuthenticatorTokenConfig
}
