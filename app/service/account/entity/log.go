package entity

// 用户登录表
type UserLoginLog struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UserID    uint64 `gorm:"column:user_id"`
	Version   string `gorm:"type:VARCHAR;column:version"`   // 客户端版本号
	Method    int8   `gorm:"type:SMALLINT;column:method"`   // 登录方式 e.g. (1: 手机号, 2:微信, 3: QQ)
	Command   int8   `gorm:"type:SMALLINT;column:command"`  // 操作类型 e.g. (1: 登录成功, 2: 登录失败, 3:登出成功, 4:登出失败)
	DeviceID  string `gorm:"type:VARCHAR;column:device_id"` // 登录时设备号
	IP        string `gorm:"type:INET;column:ip"`           // 登录 ip
}

// 用户注册表
type UserRegisterLog struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UserID    uint64 `gorm:"column:user_id"`
	Method    int8   `gorm:"type:SMALLINT;column:method"` // 注册方式 e.g. (1: 手机号)
	IP        string `gorm:"type:INET;column:ip"`         // 注册 ip
}

// 用户操作表
type UserOperationLog struct {
	ID                uint64 `gorm:"primary_key;column:id"`
	CreatedAt         int64  `gorm:"column:created_at"`
	UserID            uint64 `gorm:"column:user_id"`
	Type              int8   `gorm:"type:SMALLINT;column:type"`                // 操作类型 e.g. (1: 创建, 2: 更新, 3:删除)
	AttributeName     string `gorm:"type:SMALLINT;column:attribute_name"`      // 属性名
	AttributeOldValue string `gorm:"type:SMALLINT;column:attribute_old_value"` // 属性对应旧的值
	AttributeNewValue string `gorm:"type:SMALLINT;column:attribute_new_value"` // 属性对应新的值
}
