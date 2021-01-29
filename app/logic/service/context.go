package service

import (
	"context"
)

type clientKey struct{}

func ClientIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(clientKey{}).(string)
	return id, ok
}

func ContextWithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientKey{}, clientID)
}

func MustClientIDFromContext(ctx context.Context) string {
	if value, exists := ClientIDFromContext(ctx); exists {
		return value
	}
	panic("client id does not exist")
}
