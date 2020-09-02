package service

import (
	"context"
)

const (
	clientContextKey = "client"
)

func (s *Service) SetContextClient(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientContextKey, clientID)
}

func (s *Service) GetContext(ctx context.Context, key string) (string, bool) {
	value := ctx.Value(key)
	if value == nil {
		return "", false
	}
	id, ok := value.(string)
	if !ok {
		return "", false
	}

	return id, true
}

func (s *Service) MustGetContextClient(ctx context.Context) string {
	if value, exists := s.GetContext(ctx, clientContextKey); exists {
		return value
	}
	panic("Key \"" + clientContextKey + "\" does not exist")
}
