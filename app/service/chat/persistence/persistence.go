package persistence

import "context"

type Cacher interface {
	// Check cache
	Ping() error
	// Close cache
	Close() error
	// Add mapping
	AddMapping(uid, sid, serverID string) error
	// Set the expiration time of the mapping
	ExpireMapping(uid, sid string) (bool, error)
	// Delete the mapping
	DeleteMapping(uid, sid string) error
}

type Persister interface {
	Ping() error
	Close() error
	Client() ClientPersister
	User() UserPersister
}

type ClientPersister interface {
	GetClientCredential(ctx context.Context, id string) (string, error)

	GetClient(ctx context.Context, id string) (*Client, error)

	Create(ctx context.Context, in *ClientCreate) error

	Update(ctx context.Context, in *ClientUpdate) error

	Delete(ctx context.Context, id string) error
}

type UserPersister interface {
	Create(ctx context.Context, in *UserCreate) error

	UpdateActivated(ctx context.Context, id int64, activated bool) error

	Delete(ctx context.Context, id int64) error

	AddFriend(ctx context.Context, in *UserFriend) error

	DeleteFriend(ctx context.Context, in *UserFriend) error
}
