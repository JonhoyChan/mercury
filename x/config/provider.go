package config

import (
	"outgoing/x/log"
	"time"
)

type DefaultProvider interface {
	Logger() log.Logger
	Name() string
	Version() string
	RegisterTTL() time.Duration
	RegisterInterval() time.Duration
	Address() string
	LogMode() string
}

type RegistryProvider interface {
	Etcd() *EtcdConfig
}

type BrokerProvider interface {
	Stan() *StanConfig
}

type DatabaseProvider interface {
	Database() *DatabaseConfig
}

type RedisProvider interface {
	Redis() *RedisConfig
}

type HasherProvider interface {
	HasherArgon2() *HasherArgon2Config
	HasherBCrypt() *HasherBCryptConfig
}

type AuthenticatorProvider interface {
	AuthenticatorToken() *AuthenticatorTokenConfig
}

type GeneratorProvider interface {
	GeneratorUID() *GeneratorUIDConfig
}
