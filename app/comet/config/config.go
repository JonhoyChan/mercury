package config

import (
	"github.com/spf13/viper"
	"mercury/x/config"
)

const serviceName = "mercury-comet"

func Init() {
	v = viper.New()
	v.SetDefault(config.ViperKeyServiceName, serviceName)
	v.SetDefault(config.ViperKeyVersion, "latest")
	v.SetDefault(config.ViperKeyRegisterTTL, "30s")
	v.SetDefault(config.ViperKeyRegisterInterval, "15s")
	v.SetDefault(config.ViperKeyHost, "0.0.0.0")
	v.SetDefault(config.ViperKeyPort, 9000)

	data, err := config.LoadConfig(serviceName)
	if err != nil {
		panic("Unable to load config: " + err.Error())
	}

	for key, value := range data {
		v.Set(key, value)
	}
}
