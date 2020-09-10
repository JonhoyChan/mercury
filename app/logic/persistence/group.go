package persistence

type GroupCreate struct {
	ClientID     string
	GroupID      int64
	Name         string
	GID          string
	Introduction string
	Owner        int64
}

type GroupMember struct {
	ClientID string
	GroupID  int64
	UserID   int64
}

type Group struct {
	CreatedAt    int64
	Name         string
	GID          string
	Introduction string
	Owner        int64
	MemberCount  int64
}
