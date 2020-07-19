package entity

// 设备表
type Device struct {
	ID            uint64 `gorm:"primary_key;column:id"`
	CreatedAt     int64  `gorm:"column:created_at"`
	UpdatedAt     int64  `gorm:"column:updated_at"`
	DeviceID      string `gorm:"type:VARCHAR;column:device_id"`          // 设备ID
	Platform      uint8  `gorm:"not null;type:SMALLINT;column:platform"` // 设备平台 e.g. (1:Android, 2:IOS, 3:Web, 4:Windows, 5:MacOS)
	Brand         string `gorm:"type:VARCHAR(20);column:brand"`          // 手机厂商
	Model         string `gorm:"type:VARCHAR(20);column:model"`          // 手机厂商
	ClientVersion string `gorm:"type:VARCHAR;column:client_version"`     // 客户端版本号
	SystemVersion string `gorm:"type:VARCHAR;column:system_version"`     // 设备系统版本
	LastAt        int64  `gorm:"column:last_at"`
}
