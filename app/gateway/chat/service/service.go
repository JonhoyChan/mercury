package service

import (
	"outgoing/app/gateway/chat/session"
	"outgoing/x/log"
)

type Service struct {
	log            log.Logger
	SessionStore session.Store
}

func NewService(log log.Logger) *Service {
	return &Service{
		log:            log,
		SessionStore:  session.NewStore(),
	}
}