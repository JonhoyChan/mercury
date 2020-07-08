package service

import (
	"context"
	"outgoing/app/service/main/sms/config"
	"outgoing/x/log"

	jsoniter "github.com/json-iterator/go"

	"github.com/micro/go-micro/v2/broker"
	"github.com/panjf2000/ants/v2"
)

type Service struct {
	config config.Provider
	log    log.Logger
	pool   *ants.PoolWithFunc
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		config: c,
		log:    c.Logger(),
	}

	var err error
	s.pool, err = ants.NewPoolWithFunc(100000, s.publish)
	if err != nil {
		s.log.Warn("unable to initialize the ants pool", "error", err)
		return nil, err
	}

	return s, nil
}

func (s *Service) publish(payload interface{}) {
	message, ok := payload.(*broker.Message)
	if !ok {
		return
	}

	topic, ok := message.Header["topic"]
	if !ok {
		return
	}

	if err := broker.Publish(topic, message); err != nil {
		s.log.Error("failed to publish message", "error", err)
		return
	}
}

func (s *Service) invoke(data interface{}) error {
	body, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}

	msg := &broker.Message{
		Header: make(map[string]string),
		Body:   body,
	}

	if err := s.pool.Invoke(msg); err != nil {
		s.log.Error("failed to pool invoke", "error", err)
		return err
	}

	return nil
}

// 发送短信
func (s *Service) Send(_ context.Context, uid, country, mobile string) error {
	s.log.Info("[Send] request is received")

	return nil
}
