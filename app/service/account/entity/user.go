package entity

// 用户基础信息表
type User struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	OID       string `gorm:"type:VARCHAR;column:oid"`       // 用户唯一ID
	NickName  string `gorm:"type:VARCHAR;column:nick_name"` // 用户昵称
	Avatar    string `gorm:"type:VARCHAR;column:avatar"`
	Gender    int8   `gorm:"type:SMALLINT;column:gender"` // 设备平台 e.g. (0:未知, 1:男, 2:女)
	Bio       string `gorm:"type:VARCHAR;column:bio"`
	BirthDay  int64  `gorm:"column:birthday"` // 用户生日
	Mobile    string `gorm:"type:VARCHAR;column:mobile"`
	State     uint8  `gorm:"not null;type:SMALLINT;column:state"`
	Extra     string `gorm:"type:JSON;column:extra"`
}

// 用户状态表
type UserStatus struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	UserID    uint64 `gorm:"column:user_id"`
	Status    uint8  `gorm:"not null;type:SMALLINT;column:status"`
}

// 用户授权表
type UserAuthenticates struct {
	ID         uint64 `gorm:"primary_key;column:id"`
	CreatedAt  int64  `gorm:"column:created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at"`
	UserID     uint64 `gorm:"column:user_id"`
	Identifier string `gorm:"type:VARCHAR;column:identifier"` // 标识
	Credential string `gorm:"type:VARCHAR;column:credential"` // 密码凭证（站内的保存密码，站外的不保存或保存token）
	LastAt     int64  `gorm:"column:last_at"`
}

// 用户位置信息表
type UserLocation struct {
	ID        uint64  `gorm:"primary_key;column:id"`
	CreatedAt int64   `gorm:"column:created_at"`
	UpdatedAt int64   `gorm:"column:updated_at"`
	UserID    uint64  `gorm:"column:user_id"`
	Location  string  `gorm:"type:VARCHAR;column:location"`
	Longitude float64 `gorm:"type:DECIMAL(10,6);column:longitude"` // 经度
	Latitude  float64 `gorm:"type:DECIMAL(10,6);column:latitude"`  // 纬度
}

// 用户设备表
type UserDevices struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	UserID    uint64 `gorm:"column:user_id"`
	DeviceID  uint64 `gorm:"column:user_id"`
	Method    int8   `gorm:"type:SMALLINT;column:method"`  // 登录方式 e.g. (1: 手机号, 2:微信, 3: QQ)
	LoginAt   int64  `gorm:"column:login_at"`              // 使用该设备最新的一次登录时间
	Position  string `gorm:"type:VARCHAR;column:position"` // 使用该设备最新的一次登录地点
	IP        string `gorm:"type:INET;column:ip"`          // 使用该设备最新的一次登录 IP
}
