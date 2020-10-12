package config

import "time"

type EtcdConfig struct {
	Enable    bool
	Addresses []string
	Timeout   time.Duration
}

type StanConfig struct {
	Enable      bool
	Addresses   []string
	ClusterID   string
	DurableName string
}

type DatabaseConfig struct {
	Driver      string
	DSN         string
	Active      int
	Idle        int
	IdleTimeout time.Duration
}

type RedisConfig struct {
	Address     string
	Username    string
	Password    string
	DB          int
	IdleTimeout time.Duration
}

type HasherArgon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type HasherBCryptConfig struct {
	Cost int
}

type GeneratorIDConfig struct {
	WorkID int64
	Key    []byte
}

type AuthenticatorTokenConfig struct {
	Expire int
	Key    []byte
}
