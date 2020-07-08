package auth

import (
	"outgoing/app/service/main/auth/api"
	"outgoing/app/service/main/auth/auth/jwt"
	"outgoing/app/service/main/auth/auth/token"
	"outgoing/app/service/main/auth/persistence"
	"outgoing/x/config"
)

type Handler interface {
	// 根据 Record 生成一个新的 Token
	GenerateToken(record *api.Record) (string, error)
	// 验证 Token 并返回 Record
	Authenticate(token string) (*api.Record, string, error)
}

type HandlerType int32

const (
	Token HandlerType = iota + 1
	JWT
	finally
)

func (t HandlerType) String() string {
	switch t {
	case Token:
		return "token"
	case JWT:
		return "jwt"
	default:
		return ""
	}
}

func AllHandlerType() []HandlerType {
	handlerTypes := make([]HandlerType, 0, finally-1)
	for i := HandlerType(1); i < finally; i++ {
		handlerTypes = append(handlerTypes, i)
	}
	return handlerTypes
}

type MultiHandler map[HandlerType]Handler

func NewMultiHandler(p config.AuthenticatorProvider, c persistence.Cacher) (MultiHandler, error) {
	multiHandler := make(MultiHandler)

	for _, handlerType := range AllHandlerType() {
		var (
			h   Handler
			err error
		)
		switch handlerType {
		case Token:
			if p.AuthenticatorToken().Enable {
				h, err = token.NewAuthenticator(p)
				if err != nil {
					return nil, err
				}

				multiHandler[handlerType] = h
			}
		case JWT:
			if p.AuthenticatorJWT().Enable {
				h, err = jwt.NewAuthenticator(p, c)
				if err != nil {
					return nil, err
				}

				multiHandler[handlerType] = h
			}
		}
	}
	return multiHandler, nil
}
