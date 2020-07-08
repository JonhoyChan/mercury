package persistence

import "time"

type Cacher interface {
	Ping() error
	Close() error
	GetAuthToken(uid string) string
	SetAuthToken(uid, token string, lifetime time.Duration) error
}
