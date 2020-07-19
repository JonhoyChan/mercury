package entity

type Group struct {
	ID           uint64 `gorm:"primary_key;column:id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
	Name         string `gorm:"type:VARCHAR;column:name"`
	Introduction string `gorm:"type:VARCHAR;column:introduction"`
	Owner        uint64 `gorm:"column:owner"`
	Type         uint8  `gorm:"not null;type:SMALLINT;column:type"` // 群类型 0:普通群
	Extra        string `gorm:"type:JSON;column:extra"`
	MemberCount  int    `gorm:"column:member_count"`
}

type GroupMembers struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	GroupID   uint64 `gorm:"column:group_id"`
	UserID    uint64 `gorm:"column:user_id"`
	Label     string `gorm:"type:VARCHAR;column:label"`          // 群组内昵称
	Type      uint8  `gorm:"not null;type:SMALLINT;column:type"` // 成员类型 0:普通成员；1:管理员；2:群主
	Extra     string `gorm:"type:JSON;column:extra"`
}
