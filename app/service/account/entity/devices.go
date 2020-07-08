package entity

// 设备表
type Devices struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	UserID    uint64 `gorm:"column:user_id"`
	Hash      string `gorm:"type:VARCHAR(16);column:password"`
	DeviceID  string `gorm:"type:VARCHAR;column:device_id"`
	Platform  string `gorm:"type:VARCHAR(32);column:platform"`
	LastAt    int64  `gorm:"column:last_at"`
	Language  string `gorm:"type:VARCHAR(8);column:language"`
}
