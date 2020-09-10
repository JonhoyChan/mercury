package config

import (
	"fmt"
	"outgoing/x/config"
	"outgoing/x/log"
	"time"

	"github.com/spf13/viper"
)

const (
	viperKeyTopicPushMessage = "topic.push_message"
)

var v *viper.Viper

type ViperProvider struct {
	l log.Logger
}

func NewViperProvider() Provider {
	return &ViperProvider{
		l: log.New(log.Ctx{"service": v.GetString(config.ViperKeyServiceName)}),
	}
}

func (p *ViperProvider) Logger() log.Logger {
	return p.l
}

func (p *ViperProvider) Name() string {
	return v.GetString(config.ViperKeyServiceName)
}

func (p *ViperProvider) Version() string {
	return v.GetString(config.ViperKeyVersion)
}

func (p *ViperProvider) RegisterTTL() time.Duration {
	return v.GetDuration(config.ViperKeyRegisterTTL)
}

func (p *ViperProvider) RegisterInterval() time.Duration {
	return v.GetDuration(config.ViperKeyRegisterInterval)
}

func (p *ViperProvider) Address() string {
	return fmt.Sprintf("%s:%d", v.GetString(config.ViperKeyHost), v.GetInt(config.ViperKeyPort))
}

func (p *ViperProvider) LogMode() string {
	logMode := v.GetString(config.ViperKeyLogMode)
	// Log level(trace,debug,info,warn,error)
	if logMode == "" {
		logMode = "info"
	}
	return logMode
}

func (p *ViperProvider) Etcd() *config.EtcdConfig {
	return &config.EtcdConfig{
		Enable:    v.GetBool(config.ViperKeyEtcdEnable),
		Addresses: v.GetStringSlice(config.ViperKeyEtcdAddresses),
		Timeout:   v.GetDuration(config.ViperKeyEtcdTimeout),
	}
}

func (p *ViperProvider) Stan() *config.StanConfig {
	return &config.StanConfig{
		Enable:      v.GetBool(config.ViperKeyStanEnable),
		Addresses:   v.GetStringSlice(config.ViperKeyStanAddresses),
		ClusterID:   v.GetString(config.ViperKeyStanClusterID),
		DurableName: v.GetString(config.ViperKeyStanDurableName),
	}
}

func (p *ViperProvider) Database() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Driver:      v.GetString(config.ViperKeyDatabaseDriver),
		DSN:         v.GetString(config.ViperKeyDatabaseDSN),
		Active:      v.GetInt(config.ViperKeyDatabaseActive),
		Idle:        v.GetInt(config.ViperKeyDatabaseIdle),
		IdleTimeout: v.GetDuration(config.ViperKeyDatabaseIdleTimeout),
	}
}

func (p *ViperProvider) Redis() *config.RedisConfig {
	return &config.RedisConfig{
		Address:     v.GetString(config.ViperKeyRedisAddress),
		Username:    v.GetString(config.ViperKeyRedisUsername),
		Password:    v.GetString(config.ViperKeyRedisPassword),
		DB:          v.GetInt(config.ViperKeyRedisDB),
		IdleTimeout: v.GetDuration(config.ViperKeyRedisIdleTimeout),
	}
}

func (p *ViperProvider) AuthenticatorToken() *config.AuthenticatorTokenConfig {
	return &config.AuthenticatorTokenConfig{
		Expire: v.GetInt(config.ViperKeyAuthenticatorTokenExpire),
		Key:    []byte(v.GetString(config.ViperKeyAuthenticatorTokenKey)),
	}
}

func (p *ViperProvider) HasherArgon2() *config.HasherArgon2Config {
	return &config.HasherArgon2Config{
		Memory:      v.GetUint32(config.ViperKeyHasherArgon2Parallelism),
		Iterations:  v.GetUint32(config.ViperKeyHasherArgon2Memory),
		Parallelism: uint8(v.GetUint(config.ViperKeyHasherArgon2Iterations)),
		SaltLength:  v.GetUint32(config.ViperKeyHasherArgon2SaltLength),
		KeyLength:   v.GetUint32(config.ViperKeyHasherArgon2KeyLength),
	}
}

func (p *ViperProvider) HasherBCrypt() *config.HasherBCryptConfig {
	return &config.HasherBCryptConfig{
		Cost: v.GetInt(config.ViperKeyHasherBCryptCost),
	}
}

func (p *ViperProvider) GeneratorUID() *config.GeneratorUIDConfig {
	return &config.GeneratorUIDConfig{
		WorkID: v.GetInt64(config.ViperKeyGeneratorIDWorkID),
		Key:    []byte(v.GetString(config.ViperKeyGeneratorIDKey)),
	}
}

func (p *ViperProvider) PushMessageTopic() string {
	return v.GetString(viperKeyTopicPushMessage)
}