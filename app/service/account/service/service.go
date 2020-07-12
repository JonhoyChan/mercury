package service

import (
	"outgoing/app/service/account/config"
	"outgoing/app/service/account/entity"
	"outgoing/app/service/account/persistence"
	"outgoing/app/service/account/persistence/cache"
	"outgoing/app/service/account/persistence/sql"
	"outgoing/x/database/orm"
	"outgoing/x/database/redis"
	"outgoing/x/database/sqlx"
	"outgoing/x/log"
	"outgoing/x/password"
	"outgoing/x/types"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/pkg/errors"
)

type Service struct {
	config    config.Provider
	log       log.Logger
	persister persistence.Persister
	uidGen    types.UidGenerator
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		config: c,
		log:    c.Logger(),
	}
	err := s.withPersister()
	if err != nil {
		return nil, err
	}

	err = s.withUidGenerator()
	if err != nil {
		return nil, err
	}

	return s, nil
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
			// TODO 开发时使用，后续删除
			gormDB.AutoMigrate(
				new(entity.User),
				new(entity.UserAuth),
				new(entity.UserLocation),
				new(entity.UserLoginLog),
				new(entity.UserRegisterLog),
				new(entity.UserOperationLog),
				new(entity.Devices),
			)
			db, err := sqlx.Open(s.config)
			if err != nil {
				s.log.Error("unable to initialize the persister for sqlx, retrying", "error", err)
				return err
			}
			c, err := redis.NewClient(s.config)
			if err != nil {
				s.log.Error("unable to initialize the persister for redis, retrying", "error", err)
				return err
			}
			p := sql.NewPersister(db, cache.NewCache(c, s.log), password.NewHasherArgon2(s.config), s.log)
			if err := p.Ping(); err != nil {
				s.log.Error("unable to initialize the persister for sqlx, retrying", "error", err)
				return err
			}
			s.persister = p
			return nil
		}, bc),
	)
}

func (s *Service) withUidGenerator() error {
	if err := s.uidGen.Init(s.config); err != nil {
		s.log.Error("unable to initialize the generator for Uid", "error", err)
		return err
	}

	return nil
}

func (s *Service) Close() {
	if err := s.persister.Close(); err != nil {
		s.log.Error("unable to close the server", "error", err)
	}
}
