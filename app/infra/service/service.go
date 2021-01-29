package service

import (
	"bytes"
	"github.com/ghodss/yaml"
	"mercury/app/infra/ipfs"
	"mercury/app/infra/model"
	"mercury/config"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/secretboxer"
	"strconv"
)

type Servicer interface {
	LoadConfig() (*model.Config, error)
	AddFile(data []byte) (*model.File, error)
	CatFile(hash string) ([]byte, error)
}

type Service struct {
	boxer  *secretboxer.PassphraseBoxer
	config *config.Config
	log    log.Logger
	node   ipfs.Provider
}

func NewService(cfg *config.Config, l log.Logger) (*Service, error) {
	s := &Service{
		boxer:  secretboxer.NewPassphraseBoxer(strconv.Itoa(config.CurrentConfigRevision), secretboxer.EncodingTypeStd),
		config: cfg,
		log:    l,
	}

	var err error
	if s.node, err = ipfs.NewNode("/ip4/127.0.0.1/tcp/5001"); err != nil {
		return nil, err
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

func (s *Service) AddFile(data []byte) (*model.File, error) {
	hash, err := s.node.Add(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &model.File{Hash: hash}, nil
}

func (s *Service) CatFile(hash string) ([]byte, error) {
	data, err := s.node.Cat(hash)
	if err != nil {
		return nil, ecode.ErrDataDoesNotExist
	}
	return data, nil
}
