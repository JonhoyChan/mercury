package service

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	"github.com/micro/go-micro/v2/server"
	"github.com/pkg/errors"
	"outgoing/app/service/chat/api"
	"outgoing/app/service/chat/auth"
	"outgoing/app/service/chat/config"
	"outgoing/app/service/chat/entity"
	"outgoing/app/service/chat/persistence"
	"outgoing/app/service/chat/persistence/cache"
	"outgoing/app/service/chat/persistence/sql"
	"outgoing/x/database/orm"
	"outgoing/x/database/redis"
	"outgoing/x/database/sqlx"
	"outgoing/x/ecode"
	"outgoing/x/hash"
	"outgoing/x/log"
	"reflect"
	"time"
)

type Service struct {
	config       config.Provider
	log          log.Logger
	multiHandler auth.MultiHandler
	hash         hash.Hasher
	cache        persistence.Cacher
	persister    persistence.Persister
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		config: c,
		log:    c.Logger(),
		hash:   hash.NewHasherBCrypt(c),
	}

	err := s.withAuthMultiHandler()
	if err != nil {
		return nil, err
	}
	err = s.withCache()
	if err != nil {
		return nil, err
	}
	err = s.withPersister()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) withAuthMultiHandler() error {
	var err error
	if s.multiHandler, err = auth.NewMultiHandler(s.config); err != nil {
		s.log.Error("unable to initialize the multi handler for authentication.", "error", err)
		return err
	}

	return nil
}

func (s *Service) withCache() error {
	c, err := redis.NewClient(s.config)
	if err != nil {
		s.log.Error("unable to initialize the persister for redis, retrying", "error", err)
		return err
	}
	s.cache = cache.NewCache(c)

	return nil
}

func (s *Service) withPersister() error {
	if s.persister != nil {
		return nil
	}
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()
	return errors.WithStack(
		backoff.Retry(func() error {
			gormDB, err := orm.NewORM(s.config)
			if err != nil {
				s.log.Warn("unable to initialize the persister, retrying.", "error", err)
				return err
			}
			//TODO 开发时使用，后续删除
			gormDB.AutoMigrate(
				new(entity.Client),
				new(entity.User),
				new(entity.Friend),
				new(entity.Group),
				new(entity.GroupMember),
				new(entity.Message),
			)
			db, err := sqlx.Open(s.config)
			if err != nil {
				s.log.Error("Unable to initialize the persister for sqlx, retrying", "error", err)
				return err
			}
			p := sql.NewPersister(db)
			if err := p.Ping(); err != nil {
				s.log.Error("Unable to ping the persister, retrying", "error", err)
				return err
			}
			s.persister = p
			return nil
		}, bc),
	)
}

func (s *Service) AuthenticateClientToken(fn server.HandlerFunc) server.HandlerFunc {
	return ecode.MicroHandlerFunc(func(ctx context.Context, req server.Request, rsp interface{}) error {
		v := reflect.ValueOf(req.Body())
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
			if f := v.FieldByName("Token"); f.Kind() == reflect.String {
				var clientID string
				_, err := s.multiHandler[api.HandlerTypeDefault].Authenticate(f.String(), &clientID)
				if err != nil {
					return err
				}
				ctx = s.SetClientID(ctx, clientID)
			}
		}

		return fn(ctx, req, rsp)
	})
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
