package config

import (
	"fmt"
	"github.com/google/uuid"
	"outgoing/x/config"
	"outgoing/x/log"
	"time"

	"github.com/spf13/viper"
)

const (
	viperKeyCometServiceName      = "rpc_services.comet.service_name"
	viperKeyTopicPushMessage      = "topic.push_message"
	viperKeyTopicBroadcastMessage = "topic.broadcast_message"
)

var v *viper.Viper

type ViperProvider struct {
	id string
	l  log.Logger
}

func NewViperProvider() Provider {
	return &ViperProvider{
		id: uuid.New().String(),
		l:  log.New(log.Ctx{"service": v.GetString(config.ViperKeyServiceName)}),
	}
}

func (p *ViperProvider) ID() string {
	return p.id
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

func (p *ViperProvider) CometServiceName() string {
	return v.GetString(viperKeyCometServiceName)
}

func (p *ViperProvider) PushMessageTopic() string {
	return v.GetString(viperKeyTopicPushMessage)
}

func (p *ViperProvider) BroadcastMessageTopic() string {
	return v.GetString(viperKeyTopicBroadcastMessage)
}
