package service

import (
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/golang-lru"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"outgoing/app/infra/config"
	"outgoing/x"
	"outgoing/x/log"
	"outgoing/x/secretboxer"
	"strings"
)

var cache, _ = lru.New(16)

type Service struct {
	c     config.Provider
	boxer *secretboxer.PassphraseBoxer
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		c:     c,
		boxer: secretboxer.NewPassphraseBoxer(c.ConfigPassphrase(), secretboxer.EncodingTypeStd),
	}

	if err := s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) load() error {
	configPath := s.c.ConfigPath()
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		filename := file.Name()
		path := x.Sprintf("%s/%s", configPath, filename)

		v := viper.New()
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return err
		}

		_ = cache.Add(strings.TrimSuffix(filename, ".yml"), v)
		go s.watch(v)
	}
	return nil
}

func (s *Service) watch(v *viper.Viper) {
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Info("[watch] The configuration file has changed", "filename", e.Name)
	})
}

func (s *Service) LoadConfig(name string) (string, error) {
	var allSettings map[string]interface{}
	if v, found := cache.Get(name); found {
		allSettings = v.(*viper.Viper).AllSettings()
	} else {
		filename := name
		if !strings.HasSuffix(filename, ".yml") {
			filename += ".yml"
		}
		path := x.Sprintf("%s/%s", s.c.ConfigPath(), filename)
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}

		v := viper.New()
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return "", err
		}
		_ = cache.Add(f.Name(), v)

		allSettings = v.AllSettings()
	}

	b, err := jsoniter.Marshal(allSettings)
	if err != nil {
		return "", err
	}

	ciphertext, err := s.boxer.Seal(b)
	if err != nil {
		return "", err
	}

	return ciphertext, nil
}
