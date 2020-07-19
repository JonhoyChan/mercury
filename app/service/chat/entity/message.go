package entity

// 消息表
type Message struct {
	ID          uint64 `gorm:"primary_key;column:id"`
	CreatedAt   int64  `gorm:"column:created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
	Type        uint8  `gorm:"type:SMALLINT;column:method"` // 类型 e.g. (1: 用户, 2:群组)
	Sender      uint64 `gorm:"column:group_id"`             // 发送者id
	Receiver    uint64 `gorm:"column:message_id"`           // 接收者id,如果是单聊信息，则为user_id，如果是群组消息，则为group_id
	ContentType uint8  `gorm:"type:SMALLINT;column:method"` // 消息类型 e.g. (10: 文本, 20:图片, 30:音频, 40:文件, 50:位置, 60:红包)
	Content     string `gorm:"type:TEXT;column:content"`
	Status      uint8  `gorm:"not null;type:SMALLINT;column:status"`
	ToUserIDs   string `gorm:"type:JSON;column:to_user_ids"` // 需要@的用户id列表，多个用户用,隔开
}

// 用户消息表
type UserMessages struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	GroupID   uint64 `gorm:"column:group_id"`
	MessageID uint64 `gorm:"column:message_id"`
	Sequence  int    `gorm:"column:sequence"` // 消息序列号
}

// 群组消息表
type GroupMessages struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	GroupID   uint64 `gorm:"column:group_id"`
	MessageID uint64 `gorm:"column:message_id"`
	Sequence  int    `gorm:"column:sequence"` // 消息序列号
}
