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

	SetUserTopicLastSequence(uid, topic string, sequence int64) error

	GetUserTopicsLastSequence(uid string) (map[string]int64, error)

	SetUsersTopic(uids []string, topic string) error

	GetUserTopics(uid string) ([]string, error)
}

type Persister interface {
	Ping() error
	Close() error
	Client() ClientPersister
	User() UserPersister
	Message() MessagePersister
	Group() GroupPersister
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

	GetFriends(_ context.Context, userID int64) ([]int64, error)

	DeleteFriend(ctx context.Context, in *UserFriend) error
}

type MessagePersister interface {
	Add(_ context.Context, message *Message) error

	GetTopicLastSequence(_ context.Context, topic string) (int64, error)

	GetTopicMessageBySequence(_ context.Context, topic string, sequence int64) (*Message, error)

	GetTopicMessagesByLastSequence(_ context.Context, topic string, sequence int64) ([]*Message, int64, error)
}

type GroupPersister interface {
	Create(_ context.Context, in *GroupCreate) (*Group, error)

	AddMember(_ context.Context, in *GroupMember) error

	CheckMember(_ context.Context, groupID int64, userID int64) (bool, error)

	GetMembers(_ context.Context, clientID string, groupID int64) ([]int64, error)

	GetGroups(_ context.Context, userID int64) ([]*Group, error)
}
