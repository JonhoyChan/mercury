package service

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/server"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"outgoing/app/logic/auth/jwt"
	"outgoing/app/logic/auth/token"
	"outgoing/app/logic/config"
	"outgoing/app/logic/entity"
	"outgoing/app/logic/persistence"
	"outgoing/app/logic/persistence/cache"
	"outgoing/app/logic/persistence/sql"
	"outgoing/x/database/orm"
	"outgoing/x/database/redis"
	"outgoing/x/database/sqlx"
	"outgoing/x/ecode"
	"outgoing/x/hash"
	"outgoing/x/log"
	"outgoing/x/types"
	"reflect"
	"time"
)

type Service struct {
	config    config.Provider
	log       log.Logger
	token     token.Authenticator
	jwt       jwt.Authenticator
	hash      hash.Hasher
	cache     persistence.Cacher
	persister persistence.Persister

	idGen types.IDGenerator

	antsPool *ants.PoolWithFunc

	messageChan map[string]chan *types.Message

	doneChan chan struct{}
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		config:      c,
		log:         c.Logger(),
		hash:        hash.NewHasherBCrypt(c),
		messageChan: make(map[string]chan *types.Message),
		doneChan:    make(chan struct{}),
	}

	err := s.withTokenAuthenticator()
	if err != nil {
		return nil, err
	}
	err = s.withJWTAuthenticator()
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
	err = s.withIDGenerator()
	if err != nil {
		return nil, err
	}
	err = s.withAntsPool()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Close() error {
	if s.cache != nil {
		if err := s.cache.Close(); err != nil {
			return err
		}
	}
	if s.persister != nil {
		if err := s.persister.Close(); err != nil {
			return err
		}
	}
	if s.antsPool != nil {
		s.antsPool.Release()
	}
	s.doneChan <- struct{}{}
	close(s.doneChan)
	for _, c := range s.messageChan {
		close(c)
	}
	return nil
}

func (s *Service) withTokenAuthenticator() error {
	var err error
	if s.token, err = token.NewAuthenticator(s.config); err != nil {
		return err
	}
	return nil
}

func (s *Service) withJWTAuthenticator() error {
	var err error
	if s.jwt, err = jwt.NewAuthenticator(); err != nil {
		return err
	}
	return nil
}

func (s *Service) withCache() error {
	c, err := redis.NewClient(s.config)
	if err != nil {
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
				return err
			}
			p := sql.NewPersister(db)
			if err := p.Ping(); err != nil {
				s.log.Error("unable to ping the persister, retrying", "error", err)
				return err
			}
			s.persister = p
			return nil
		}, bc),
	)
}

func (s *Service) withIDGenerator() error {
	if err := s.idGen.Init(s.config); err != nil {
		return err
	}
	return nil
}

func (s *Service) withAntsPool() error {
	var err error
	s.antsPool, err = ants.NewPoolWithFunc(100000, s.publish)
	if err != nil {
		return err
	}
	return nil
}

type PublishMessage struct {
	Topic   string
	Message *broker.Message
}

func (s *Service) publish(payload interface{}) {
	publicMessage, ok := payload.(*PublishMessage)
	if !ok {
		return
	}

	if err := broker.Publish(publicMessage.Topic, publicMessage.Message); err != nil {
		s.log.Error("failed to publish message", "error", err)
		return
	}
}

type marshaler interface {
	Marshal() ([]byte, error)
}

func (s *Service) invoke(topic string, m marshaler) error {
	body, err := m.Marshal()
	if err != nil {
		return err
	}

	pm := &PublishMessage{
		Topic: topic,
		Message: &broker.Message{
			Header: make(map[string]string),
			Body:   body,
		},
	}

	if err := s.antsPool.Invoke(pm); err != nil {
		s.log.Error("failed to pool invoke", "error", err)
		return err
	}

	return nil
}

func (s *Service) AuthenticateClientToken(fn server.HandlerFunc) server.HandlerFunc {
	return ecode.MicroHandlerFunc(func(ctx context.Context, req server.Request, rsp interface{}) error {
		if !req.Stream() {
			v := reflect.ValueOf(req.Body())
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
				if f := v.FieldByName("Token"); f.Kind() == reflect.String {
					var clientID string
					_, err := s.token.Authenticate(f.String(), &clientID)
					if err != nil {
						s.log.Error("[AuthenticateClientToken] failed to authenticating the token", "error", err)
						return ecode.ErrInvalidToken
					}
					ctx = s.SetContextClient(ctx, clientID)
				}
			}
		}

		return fn(ctx, req, rsp)
	})
}

func (s *Service) DecodeID(uid types.ID) int64 {
	return s.idGen.DecodeID(uid)
}

func (s *Service) EncodeID(val int64) types.ID {
	return s.idGen.EncodeInt64(val)
}
