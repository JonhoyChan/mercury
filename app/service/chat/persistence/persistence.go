package persistence

import (
	"context"
	"time"
)

type Cacher interface {
	Ping() error

	Close() error

	AddMapping(uid, sid, serverID string) error

	ExpireMapping(uid, sid string) (bool, error)

	DeleteMapping(uid, sid string) error

	GetSessions(uids ...string) (map[string]string, []string, error)

	GetServerIDs(sids ...string) ([]string, error)

	GetClient(clientID string) (*Client, error)

	SetClient(clientID string, client *Client) error

	DeleteClient(clientID string) error

	GetClientID(token string) string

	SetClientID(token, clientID string, lifetime time.Duration) error

	GetTopicSequence(topic string) (int64, error)

	SetTopicSequence(topic string, sequence int64, lifetime time.Duration) error

	IncrTopicSequence(topic string) (int64, error)
}

type Persister interface {
	Ping() error
	Close() error
	Client() ClientPersister
	User() UserPersister
	Message() MessagePersister
}

type ClientPersister interface {
	GetClientCredential(ctx context.Context, id string) (string, error)

	GetClient(ctx context.Context, id string) (*Client, error)

	Create(ctx context.Context, in *ClientCreate) error

	Update(ctx context.Context, in *ClientUpdate) error

	Delete(ctx context.Context, id string) error
}

type UserPersister interface {
	CheckActivated(_ context.Context, clientID, uid string) (bool, error)

	Create(ctx context.Context, in *UserCreate) error

	UpdateActivated(ctx context.Context, id int64, activated bool) error

	Delete(ctx context.Context, id int64) error

	AddFriend(ctx context.Context, in *UserFriend) error

	DeleteFriend(ctx context.Context, in *UserFriend) error
}

type MessagePersister interface {
	Add(_ context.Context, message *Message) error

	GetTopicLastSequence(_ context.Context, topic string) (int64, error)
}
