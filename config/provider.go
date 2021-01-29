package config

var _ Provider = new(ProviderConfig)

type Provider interface {
	GetService(name string) (*Service, bool)
	Revision() int
	LogLevel() string
	Services() []*Service
	Registry() *Registry
	Broker() *Broker
	Database() *Database
	Redis() *Redis
	Authenticator() *Authenticator
	Hasher() *Hasher
	Generator() *Generator
	Topic() Topic
}

type ProviderConfig struct {
	*Config
}

func (p *ProviderConfig) Revision() int {
	return p.Config.Revision
}

func (p *ProviderConfig) LogLevel() string {
	return p.Config.LogLevel
}

func (p *ProviderConfig) Services() []*Service {
	return p.Config.Services
}

func (p *ProviderConfig) Registry() *Registry {
	return p.Config.Registry
}

func (p *ProviderConfig) Broker() *Broker {
	return p.Config.Broker
}

func (p *ProviderConfig) Database() *Database {
	return p.Config.Database
}

func (p *ProviderConfig) Redis() *Redis {
	return p.Config.Redis
}

func (p *ProviderConfig) Authenticator() *Authenticator {
	return p.Config.Authenticator
}

func (p *ProviderConfig) Hasher() *Hasher {
	return p.Config.Hasher
}

func (p *ProviderConfig) Generator() *Generator {
	return p.Config.Generator
}

func (p *ProviderConfig) Topic() Topic {
	return p.Config.Topic
}

func NewProviderConfig(cfg *Config) *ProviderConfig {
	return &ProviderConfig{cfg}
}
