package auth

import (
	"outgoing/app/service/chat/api"
	"outgoing/app/service/chat/auth/token"
	"outgoing/x/config"
)

type (
	Handler interface {
		GenerateToken(data interface{}) (token string, lifetime string, err error)

		Authenticate(token string, out interface{}) (lifetime string, err error)
	}
	authenticatorProvider interface {
		AuthenticatorToken() *config.AuthenticatorTokenConfig
	}
)

func AllHandlerType() []api.HandlerType {
	handlerTypes := make([]api.HandlerType, 0, len(api.HandlerType_name))
	for i, _ := range api.HandlerType_name {
		handlerTypes = append(handlerTypes, api.HandlerType(i))
	}
	return handlerTypes
}

type MultiHandler map[api.HandlerType]Handler

func NewMultiHandler(p authenticatorProvider) (MultiHandler, error) {
	multiHandler := make(MultiHandler)

	for _, handlerType := range AllHandlerType() {
		var (
			h   Handler
			err error
		)
		switch handlerType {
		case api.HandlerTypeDefault:
			h, err = token.NewAuthenticator(p)
			if err != nil {
				return nil, err
			}

			multiHandler[handlerType] = h
		}
	}
	return multiHandler, nil
}
