package persistence

type UserCreate struct {
	ClientID string
	UserID   int64
	Name     string
	UID      string
}

type UserFriend struct {
	ClientID     string
	UserID       int64
	FriendUserID int64
}
