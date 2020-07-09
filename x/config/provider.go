package config

import (
	"outgoing/x/log"
	"time"
)

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
}

type GeneratorProvider interface {
	GeneratorUid() *GeneratorUidConfig
}

type AuthenticatorProvider interface {
	AuthenticatorToken() *AuthenticatorTokenConfig
	AuthenticatorJWT() *AuthenticatorJWTConfig
}

type DefaultProvider interface {
	ID() string
	Logger() log.Logger
	Name() string
	Version() string
	RegisterTTL() time.Duration
	RegisterInterval() time.Duration
	Address() string
	LogMode() string
}
