package service

import (
	"context"
	"outgoing/app/service/main/auth/api"
	"outgoing/app/service/main/auth/auth"
	"outgoing/app/service/main/auth/config"
	"outgoing/app/service/main/auth/persistence/cache"
	"outgoing/x/database/redis"
	"outgoing/x/ecode"
	"outgoing/x/log"
)

type Service struct {
	config       config.Provider
	log          log.Logger
	multiHandler auth.MultiHandler
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		config: c,
		log:    c.Logger(),
	}

	err := s.withAuthMultiHandler()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) withAuthMultiHandler() error {
	c, err := redis.NewClient(s.config)
	if err != nil {
		s.log.Error("unable to initialize the persister for redis, retrying", "error", err)
		return err
	}

	if s.multiHandler, err = auth.NewMultiHandler(s.config, cache.NewCache(c, s.log)); err != nil {
		s.log.Error("unable to initialize the multi handler for authentication.", "error", err)
		return err
	}

	return nil
}

// 根据 Record 生成一个新的 Token
func (s *Service) GenerateToken(_ context.Context, t auth.HandlerType, record *api.Record) (string, error) {
	s.log.Info("[GenerateToken] request is received")

	token, err := s.multiHandler[t].GenerateToken(record)
	if err != nil {
		return "", ecode.Wrap(err, "failed to generate token")
	}

	return token, nil
}

// 验证 Token 并返回 Record
func (s *Service) Authenticate(_ context.Context, t auth.HandlerType, token string) (*api.Record, error) {
	s.log.Info("[Authenticate] request is received")

	record, token, err := s.multiHandler[t].Authenticate(token)
	if err != nil {
		return nil, ecode.Wrap(err, "failed to generate token")
	}

	return record, nil
}
