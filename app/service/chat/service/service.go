package service

import (
	"context"
	"outgoing/app/service/chat/config"
	"outgoing/app/service/chat/persistence"
	"outgoing/app/service/chat/persistence/cache"
	"outgoing/x/database/redis"
	"outgoing/x/log"
)

type Service struct {
	config config.Provider
	log    log.Logger
	cache  persistence.Cacher
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
	s.cache = cache.NewCache(c, s.log)

	return nil
}

// Connect a connection
func (s *Service) Connect(_ context.Context, uid, sid, serverID string) error {
	s.log.Info("[Connect] request is received")

	if err := s.cache.AddMapping(uid, sid, serverID); err != nil {
		s.log.Error("[Connect] failed to add mapping", "uid", uid, "error", err)
		return err
	}

	return nil
}

// Disconnect a connection
func (s *Service) Disconnect(_ context.Context, uid, sid string) error {
	s.log.Info("[Disconnect] request is received")

	if err := s.cache.DeleteMapping(uid, sid); err != nil {
		s.log.Error("[Disconnect] failed to delete mapping", "uid", uid, "error", err)
		return err
	}

	return nil
}

// Heartbeat a connection
func (s *Service) Heartbeat(_ context.Context, uid, sid, serverID string) error {
	s.log.Info("[Heartbeat] request is received")

	expired, err := s.cache.ExpireMapping(uid, sid)
	if err != nil {
		s.log.Error("[Heartbeat] failed to expire mapping", "uid", uid, "error", err)
		return err
	}
	if !expired {
		if err := s.cache.AddMapping(uid, sid, serverID); err != nil {
			s.log.Error("[Heartbeat] failed to add mapping", "uid", uid, "error", err)
			return err
		}
	}

	return nil
}
