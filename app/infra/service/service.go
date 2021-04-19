package service

import (
	"github.com/ghodss/yaml"
	"mercury/app/infra/model"
	"mercury/config"
	"mercury/x/log"
	"mercury/x/secretboxer"
	"strconv"
)

type Servicer interface {
	LoadConfig() (*model.Config, error)
}

type Service struct {
	boxer  *secretboxer.PassphraseBoxer
	config *config.Config
	log    log.Logger
}

func NewService(cfg *config.Config, l log.Logger) (*Service, error) {
	s := &Service{
		boxer:  secretboxer.NewPassphraseBoxer(strconv.Itoa(config.CurrentConfigRevision), secretboxer.EncodingTypeStd),
		config: cfg,
		log:    l,
	}

	return s, nil
}

func (s *Service) LoadConfig() (*model.Config, error) {
	b, err := yaml.Marshal(s.config)
	if err != nil {
		return nil, err
	}

	ciphertext, err := s.boxer.Seal(b)
	if err != nil {
		return nil, err
	}

	return &model.Config{Ciphertext: ciphertext}, nil
}
