package config

import (
	"fmt"
	"outgoing/x/config"
	"outgoing/x/log"
	"time"

	"github.com/spf13/viper"
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
