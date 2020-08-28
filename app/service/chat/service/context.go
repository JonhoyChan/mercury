package service

import "context"

const (
	clientContextKey = "client"
	userContextKey   = "user"
)

func (s *Service) SetContextClient(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientContextKey, clientID)
}

type ContextUser struct {
	ID  int64
	UID string
}

func (s *Service) SetContextUser(ctx context.Context, UID string) context.Context {
	u := &ContextUser{
		ID:  s.DecodeUid(UID),
		UID: UID,
	}
	return context.WithValue(ctx, userContextKey, u)
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

func (s *Service) GetContextUser(ctx context.Context, key string) (*ContextUser, bool) {
	value := ctx.Value(key)
	if value == nil {
		return nil, false
	}
	u, ok := value.(*ContextUser)
	if !ok {
		return nil, false
	}

	return u, true
}

func (s *Service) MustGetContextClient(ctx context.Context) string {
	if value, exists := s.GetContext(ctx, clientContextKey); exists {
		return value
	}
	panic("Key \"" + clientContextKey + "\" does not exist")
}

func (s *Service) MustGetContextUser(ctx context.Context) *ContextUser {
	if value, exists := s.GetContextUser(ctx, userContextKey); exists {
		return value
	}
	panic("Key \"" + userContextKey + "\" does not exist")
}
