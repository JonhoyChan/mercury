package model

import (
	"outgoing/x/types"
	"time"
)

type User struct {
	Uid       types.Uid `json:"uid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	NickName  string    `json:"nick_name"`
	Avatar    string    `json:"avatar"`
	Gender    int32     `json:"gender"`
	Bio       string    `json:"bio"`
	Birthday  string    `json:"birthday"`
	Mobile    string    `json:"mobile"`
	State     int32     `json:"state"`
}

type Topic struct {
	ID        types.Uid
	CreatedAt time.Time
	UpdatedAt time.Time

	// State of the topic: normal (ok), suspended, deleted
	State   int32
	StateAt *time.Time

	// Timestamp when the last message has passed through the topic
	TouchedAt time.Time

	// Topic owner. Could be zero
	Owner string

	// Server-issued sequential ID
	SeqID int
	// If messages were deleted, sequential id of the last operation to delete them
	DelID int
}

// Subscription to a topic
type Subscription struct {
	ID        types.Uid
	CreatedAt time.Time
	UpdatedAt time.Time
	// User who has relationship with the topic
	User string
	// Topic subscribed to
	Topic     string
	DeletedAt *time.Time

	// Values persisted through subscription soft-deletion

	// ID of the latest Soft-delete operation
	DelId int
	// Last SeqId reported by user as received by at least one of his sessions
	RecvSeqId int
	// Last SeqID reported read by the user
	ReadSeqId int

	// Access mode requested by this user
	//ModeWant AccessMode
	// Access mode granted to this user
	//ModeGiven AccessMode

	// Deserialized SeqID from user or topic
	seqId int
	// Deserialized TouchedAt from topic
	touchedAt time.Time
	// timestamp when the user was last online
	lastOnline time.Time
	// user agent string of the last online access
	userAgent string

	// P2P only. ID of the other user
	with string
	// P2P only. Default access: this is the mode given by the other user to this user
	//modeDefault *DefaultAccess

	// Topic's or user's state.
	state int32
}

// SetWith sets other user for P2P subscriptions.
func (s *Subscription) SetWith(with string) {
	s.with = with
}

// GetWith returns the other user for P2P subscriptions.
func (s *Subscription) GetWith() string {
	return s.with
}

// GetTouchedAt returns touchedAt.
func (s *Subscription) GetTouchedAt() time.Time {
	return s.touchedAt
}

// SetTouchedAt sets the value of touchedAt.
func (s *Subscription) SetTouchedAt(touchedAt time.Time) {
	if touchedAt.After(s.touchedAt) {
		s.touchedAt = touchedAt
	}

	if s.touchedAt.Before(s.UpdatedAt) {
		s.touchedAt = s.UpdatedAt
	}
}

// GetSeqId returns seqId.
func (s *Subscription) GetSeqId() int {
	return s.seqId
}

// SetSeqId sets seqId field.
func (s *Subscription) SetSeqId(id int) {
	s.seqId = id
}

// GetLastSeen returns lastSeen.
func (s *Subscription) GetLastOnline() time.Time {
	return s.lastOnline
}

// GetUserAgent returns userAgent.
func (s *Subscription) GetUserAgent() string {
	return s.userAgent
}

// SetLastOnlineAndUA updates lastOnline time and userAgent.
func (s *Subscription) SetLastOnlineAndUA(when *time.Time, ua string) {
	if when != nil {
		s.lastOnline = *when
	}
	s.userAgent = ua
}

// GetState returns topic's or user's state.
func (s *Subscription) GetState() int32 {
	return s.state
}

// SetState assigns topic's or user's state.
func (s *Subscription) SetState(state int32) {
	s.state = state
}
