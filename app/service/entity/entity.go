package entity

import (
	"outgoing/x/types"
)

type Client struct {
	ID          string `gorm:"primary_key;type:UUID;column:id"`
	CreatedAt   int64  `gorm:"column:created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
	Name        string `gorm:"not null;type:VARCHAR;column:name"`
	TokenSecret string `gorm:"not null;type:VARCHAR;column:token_secret"`
	TokenExpire int64  `gorm:"type:VARCHAR;column:token_expire"`
	Credential  string `gorm:"type:VARCHAR;column:credential"`
	UserCount   int32  `gorm:"column:user_count"`
	GroupCount  int32  `gorm:"column:group_count"`
}

type User struct {
	ID        int64  `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	ClientID  string `gorm:"type:VARCHAR;column:client_id"`
	Name      string `gorm:"type:VARCHAR;column:name"`
	// Unique ID of the user.
	UID       string `gorm:"not null;type:VARCHAR;column:uid"`
	Activated bool   `gorm:"default:true;column:activated"`
}

type Friend struct {
	ID           uint64 `gorm:"primary_key;column:id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
	UserID       int64  `gorm:"column:user_id"`
	FriendUserID int64  `gorm:"column:friend_user_id"`
}

type Group struct {
	ID        int64  `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	ClientID  string `gorm:"type:VARCHAR;column:client_id"`
	Name      string `gorm:"type:VARCHAR;column:name"`
	// Unique ID of the group.
	GID          string `gorm:"not null;type:VARCHAR;column:gid"`
	Introduction string `gorm:"type:VARCHAR;column:introduction"`
	Owner        int64  `gorm:"column:owner"`
	// Group type e.g. (0: public group)
	Type        uint8 `gorm:"not null;type:SMALLINT;column:type"`
	Activated   bool  `gorm:"default:true;column:activated"`
	MemberCount int32 `gorm:"column:member_count"`
}

type GroupMember struct {
	ID        int64 `gorm:"primary_key;column:id"`
	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
	GroupID   int64 `gorm:"column:group_id"`
	UserID    int64 `gorm:"column:user_id"`
}

type Message struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	// The topic of the message
	Topic string `gorm:"type:VARCHAR;column:topic"`
	// The sequence of the topic
	Sequence int64 `gorm:"column:sequence"`
	// Message type. e.g. (1: single chat message, 2:group chat message)
	MessageType types.MessageType `gorm:"type:SMALLINT;column:message_type"`
	// Message sender
	Sender int64 `gorm:"column:sender"`
	// Message receiver
	// If a single chat, the value is user ID.
	// If a group chat, the value is group ID.
	Receiver int64 `gorm:"column:receiver"`
	// Message content type. e.g. {"content": "Hello, World!". "content_type": 10}
	ContentType types.ContentType `gorm:"type:SMALLINT;column:content_type"`
	// The body of the message, will change according to the content type
	Body string `gorm:"type:JSON;column:body"`
	// The status of the message. e.g. (0: normal, 1:recalled, 3:deleted)
	Status uint8 `gorm:"not null;type:SMALLINT;column:status"`
	// List of user IDs mentioned in the message
	Mentions string `gorm:"column:mentions"`
}
