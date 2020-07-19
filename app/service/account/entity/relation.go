package entity

type UserFriends struct {
	ID           uint64 `gorm:"primary_key;column:id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
	UserID       uint64 `gorm:"column:user_id"`
	FriendUserID uint64 `gorm:"column:friend_user_id"`
	Label        string `gorm:"type:VARCHAR;column:label"` // 朋友备注
	Extra        string `gorm:"type:JSON;column:extra"`
	State        uint8  `gorm:"not null;type:SMALLINT;column:state"` // 状态 0:正常；1:已删除；2:黑名单
}

type FriendRequest struct {
	ID           uint64 `gorm:"primary_key;column:id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
	UserID       uint64 `gorm:"column:user_id"`
	FriendUserID uint64 `gorm:"column:friend_user_id"`
	Reason       string `gorm:"type:VARCHAR;column:reason"`
	Status       uint8  `gorm:"not null;type:SMALLINT;column:status"`
}
