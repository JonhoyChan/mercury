package entity

// 用户基础信息表
type User struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	VID       string `gorm:"type:VARCHAR;column:vid"`       // 用户唯一ID
	NickName  string `gorm:"type:VARCHAR;column:nick_name"` // 用户昵称
	Avatar    string `gorm:"type:VARCHAR;column:avatar"`
	Gender    int8   `gorm:"type:SMALLINT;column:gender"`
	Bio       string `gorm:"type:VARCHAR;column:bio"`
	BirthDay  int64  `gorm:"column:birthday"` // 用户生日
	Mobile    string `gorm:"type:VARCHAR;column:mobile"`
	State     uint8  `gorm:"not null;type:SMALLINT;column:state"`
}

// 用户授权表
type UserAuth struct {
	ID           uint64 `gorm:"primary_key;column:id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
	UserID       uint64 `gorm:"column:user_id"`
	IdentityType int8   `gorm:"not null;type:SMALLINT;column:identity_type"` // 登录类型 e.g. (1: 手机号, 2:VID)
	Identifier   string `gorm:"type:VARCHAR;column:identifier"`              // 标识
	Credential   string `gorm:"type:VARCHAR;column:credential"`              // 密码凭证（站内的保存密码，站外的不保存或保存token）
	LastAt       int64  `gorm:"column:last_at"`
}

// 用户位置信息表
type UserLocation struct {
	UserID    uint64  `gorm:"column:user_id"`
	CreatedAt int64   `gorm:"column:created_at"`
	UpdatedAt int64   `gorm:"column:updated_at"`
	Location  string  `gorm:"type:VARCHAR;column:location"`
	Longitude float64 `gorm:"type:DECIMAL(10,6);column:longitude"` // 经度
	Latitude  float64 `gorm:"type:DECIMAL(10,6);column:latitude"`  // 纬度
}

// 用户设备表
type UserDevices struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UserID    uint64 `gorm:"column:user_id"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	Version   string `gorm:"type:VARCHAR;column:version"`    // 客户端版本号
	Method    int8   `gorm:"type:SMALLINT;column:method"`    // 登录方式 e.g. (1: 手机号, 2:微信, 3: QQ)
	DeviceID  string `gorm:"type:VARCHAR;column:device_id"`  // 登录设备
	LoginAt   int64  `gorm:"column:login_at"`                // 使用该设备最新的一次登录时间
	Position  string `gorm:"type:VARCHAR;column:position"`   // 使用该设备最新的一次登录地点
	IP        string `gorm:"type:INET;column:ip"`            // 使用该设备最新的一次登录 IP
	OS        string `gorm:"type:VARCHAR;column:os"`         // 设备系统
	OSVersion string `gorm:"type:VARCHAR;column:os_version"` // 设备系统版本
}
