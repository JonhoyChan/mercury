package config

import (
	"outgoing/x/config"

	"github.com/spf13/viper"
)

func Init(configFile string) {
	v = viper.New()

	// enable ability to specify configuration file via flag
	v.SetConfigFile(configFile)

	v.SetDefault(config.ViperKeyServiceName, "service.chat.logic")
	v.SetDefault(config.ViperKeyVersion, "latest")
	v.SetDefault(config.ViperKeyRegisterTTL, "30s")
	v.SetDefault(config.ViperKeyRegisterInterval, "15s")
	v.SetDefault(config.ViperKeyHost, "0.0.0.0")
	v.SetDefault(config.ViperKeyPort, 10001)

	// If a configuration file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		panic("unable to found configuration file:" + err.Error())
	}
}
