package service

import "context"

const (
	clientContextKey = "client"
)

func (s *Service) SetClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientContextKey, clientID)
}

func (s *Service) GetClientID(ctx context.Context) (string, bool) {
	value := ctx.Value(clientContextKey)
	if value == nil {
		return "", false
	}
	id, ok := value.(string)
	if !ok {
		return "", false
	}

	return id, true
}

func (s *Service) MustGetClientID(ctx context.Context) string {
	if value, exists := s.GetClientID(ctx); exists {
		return value
	}
	panic("Key \"" + clientContextKey + "\" does not exist")
}
