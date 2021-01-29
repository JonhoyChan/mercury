package config

import (
	"context"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"mercury/config/api"
	"mercury/x/fill"
	"mercury/x/secretboxer"
	"strconv"
)

// CurrentConfigRevision is the latest configuration revision configurations
// that don't match this revision number should be migrated up
const CurrentConfigRevision = 2

// Log level(trace,debug,info,warn,error)
const DefaultLogLevel = "debug"

// Config encapsulates all configuration details for mercury
type Config struct {
	path string

	Revision      int            `json:"revision"`
	LogLevel      string         `json:"log_level"`
	Services      []*Service     `json:"services"`
	Registry      *Registry      `json:"registry"`
	Broker        *Broker        `json:"broker"`
	Database      *Database      `json:"database"`
	Redis         *Redis         `json:"redis"`
	Authenticator *Authenticator `json:"authenticator"`
	Hasher        *Hasher        `json:"hasher"`
	Generator     *Generator     `json:"generator"`
	Topic         Topic          `json:"topic"`
}

func (cfg Config) GetService(name string) (*Service, bool) {
	for i := 0; i < len(cfg.Services); i++ {
		if cfg.Services[i].Name == name {
			return cfg.Services[i], true
		}
	}

	return nil, false
}

// WriteToFile encodes a configration to YAML and writes it to path
func (cfg Config) WriteToFile(path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 06777)
}

// ReadFromFile reads a YAML configuration file from path
func ReadFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]interface{})
	if err = yaml.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	cfg := &Config{path: path}

	if rev, ok := fields["revision"]; ok {
		cfg.Revision = (int)(rev.(float64))
	}
	if err = fill.Struct(fields, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func ReadFromRemote(ctx context.Context, url string) (*Config, error) {
	resp, err := api.New(url).LoadConfig(ctx)
	if err != nil {
		return nil, err
	}

	boxer := secretboxer.NewPassphraseBoxer(strconv.Itoa(CurrentConfigRevision), secretboxer.EncodingTypeStd)
	data, err := boxer.Open(resp.Ciphertext)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Revision:      CurrentConfigRevision,
		LogLevel:      DefaultLogLevel,
		Services:      DefaultServices(),
		Registry:      DefaultRegistry(),
		Broker:        DefaultBroker(),
		Database:      DefaultDatabase(),
		Redis:         DefaultRedis(),
		Authenticator: DefaultAuthenticator(),
		Hasher:        DefaultHasher(),
		Generator:     DefaultGenerator(),
		Topic:         DefaultTopic(),
	}
}
