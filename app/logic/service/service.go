package service

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"
	"github.com/pkg/errors"
	"mercury/app/logic/api"
	"mercury/app/logic/auth/jwt"
	"mercury/app/logic/auth/token"
	"mercury/app/logic/persistence"
	"mercury/app/logic/persistence/cache"
	"mercury/app/logic/persistence/sql"
	"mercury/config"
	"mercury/x/database/redis"
	"mercury/x/database/sqlx"
	"mercury/x/ecode"
	"mercury/x/hash"
	"mercury/x/log"
	"mercury/x/types"
	"reflect"
	"strings"
	"time"
)

type Servicer interface {
	Authenticate(token string, out interface{}) (string, error)

	GetClient(ctx context.Context) (*api.Client, error)
	CreateClient(ctx context.Context, req *api.CreateClientReq) (string, string, error)
	UpdateClient(ctx context.Context, req *api.UpdateClientReq) error
	DeleteClient(ctx context.Context) error
	GenerateToken(ctx context.Context, req *api.GenerateTokenReq) (string, string, error)
	Listen(ctx context.Context, token string, stream api.ChatClientAdmin_ListenStream) error

	Connect(ctx context.Context, req *api.ConnectReq) (string, string, error)
	Disconnect(ctx context.Context, req *api.DisconnectReq) error
	Heartbeat(ctx context.Context, req *api.HeartbeatReq) error

	CreateGroup(ctx context.Context, req *api.CreateGroupReq) (*api.Group, error)
	GetGroups(ctx context.Context, uid string) ([]*api.Group, error)
	AddMember(ctx context.Context, req *api.AddMemberReq) error
	GetMembers(ctx context.Context, gid string) ([]string, error)

	PushMessage(ctx context.Context, req *api.PushMessageReq) (int64, int64, error)
	PullMessage(ctx context.Context, req *api.PullMessageReq) ([]*api.TopicMessages, error)
	ReadMessage(ctx context.Context, req *api.ReadMessageReq) error
	Keypress(ctx context.Context, req *api.KeypressReq) error

	CreateUser(ctx context.Context, req *api.CreateUserReq) (string, error)
	UpdateActivated(ctx context.Context, uid string, activated bool) error
	DeleteUser(ctx context.Context, uid string) error
	AddFriend(ctx context.Context, uid, friendUID string) error
	DeleteFriend(ctx context.Context, uid, friendUID string) error
	GenerateUserToken(ctx context.Context, uid string) (string, string, error)
	GetFriends(ctx context.Context, uid string) ([]string, error)
}

type Service struct {
	config    config.Provider
	log       log.Logger
	token     token.Authenticator
	jwt       jwt.Authenticator
	hash      hash.Hasher
	cache     persistence.Cacher
	persister persistence.Persister

	idGen types.IDGenerator

	messageChan       map[string]chan *types.Message
	brokerMessageChan chan *PublishMessage
	doneChan          chan struct{}
}

func NewService(c config.Provider, l log.Logger) (*Service, error) {
	s := &Service{
		config:            c,
		log:               l,
		hash:              hash.NewHasherBCrypt(c),
		messageChan:       make(map[string]chan *types.Message),
		brokerMessageChan: make(chan *PublishMessage, 4096),
		doneChan:          make(chan struct{}),
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

	go s.process()
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
	s.doneChan <- struct{}{}
	close(s.doneChan)
	for _, c := range s.messageChan {
		close(c)
	}
	return nil
}

func (s *Service) process() {
	for {
		select {
		case m := <-s.brokerMessageChan:
			if err := broker.Publish(m.Topic, m.Message); err != nil {
				s.log.Error("failed to publish message", "error", err)
				return
			}
		case <-s.doneChan:
			close(s.brokerMessageChan)
			if err := broker.Disconnect(); err != nil {
				s.log.Warn("failed to disconnecting broker", "error", err)
			}
		}
	}
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
			//gormDB, err := orm.NewORM(s.config)
			//if err != nil {
			//	return err
			//}
			////TODO 开发时使用，后续删除
			//gormDB.AutoMigrate(
			//	new(entity.Client),
			//	new(entity.User),
			//	new(entity.Friend),
			//	new(entity.Group),
			//	new(entity.GroupMember),
			//	new(entity.Message),
			//)
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
	if err := s.idGen.Init(s.config.Generator().ID); err != nil {
		return err
	}
	return nil
}

type PublishMessage struct {
	Topic   string
	Message *broker.Message
}

type marshaler interface {
	Marshal() ([]byte, error)
}

func (s *Service) invoke(topic string, m marshaler) error {
	body, err := m.Marshal()
	if err != nil {
		return err
	}

	s.brokerMessageChan <- &PublishMessage{
		Topic: topic,
		Message: &broker.Message{
			Body: body,
		},
	}
	return nil
}

func AuthenticateClientToken(srv Servicer) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return ecode.MicroHandlerFunc(func(ctx context.Context, req server.Request, rsp interface{}) error {
			if !req.Stream() {
				if strings.HasPrefix(req.Endpoint(), "ChatAdmin.") {
					//if sign, ok := metadata.Get(ctx, "Sign"); ok {
					//	// TODO check sign
					//	fmt.Println(sign)
					//
					timestamp, _ := metadata.Get(ctx, "Timestamp")
					issuer, _ := metadata.Get(ctx, "Issuer")
					fmt.Println(timestamp, issuer)

					clientID, _ := metadata.Get(ctx, "Id")
					ctx = ContextWithClientID(ctx, clientID)
				} else {
					v := reflect.ValueOf(req.Body())
					if v.Kind() == reflect.Ptr {
						v = v.Elem()
						if f := v.FieldByName("Token"); f.Kind() == reflect.String {
							var clientID string
							_, err := srv.Authenticate(f.String(), &clientID)
							if err != nil {
								return ecode.ErrInvalidToken
							}
							ctx = ContextWithClientID(ctx, clientID)
						}
					}
				}
			}

			//s.log.Info("[Authenticate] new request", "endpoint", req.Endpoint())
			return fn(ctx, req, rsp)
		})
	}
}

func (s *Service) Authenticate(token string, out interface{}) (lifetime string, err error) {
	lifetime, err = s.token.Authenticate(token, out)
	if err != nil {
		s.log.Error("[AuthenticateClientToken] failed to authenticating the token", "error", err)
		return
	}

	return
}

func (s *Service) DecodeID(uid types.ID) int64 {
	return s.idGen.DecodeID(uid)
}

func (s *Service) EncodeID(val int64) types.ID {
	return s.idGen.EncodeInt64(val)
}
