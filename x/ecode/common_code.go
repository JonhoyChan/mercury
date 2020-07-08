package ecode

var (
	Success      = add(200, "success")
	ResetContent = add(205, "reset content")

	// 错误的请求
	ErrBadRequest = add(400, "bad request")
	// 权限不足
	ErrUnauthorized = add(401, "unauthorized")
	// 拒绝执行
	ErrForbidden = add(403, "forbidden")
	// 资源找不到
	ErrNotFound = add(404, "not found")
	// 请求方法不支持
	ErrMethodNotAllowed = add(405, "method not allowed")
	// 服务请求超时
	ErrRequestTimeout = add(408, "request timeout")
	// 资源被锁定
	ErrLocked = add(423, "locked")
	// 请求过于频繁
	ErrTooManyRequests = add(429, "too many requests")
	// 服务内部错误
	ErrInternalServer = add(500, "internal server")

	// 无效的Token
	ErrInvalidToken = add(1001, "invalid token")
	// Token已过期
	ErrTokenExpired = add(1002, "token has expired")
	// 异常
	ErrMalformed = add(1003, "malformed")
	// 数据已存在
	ErrDataAlreadyExist = add(1004, "data already exist")
	// 数据不存在
	ErrDataDoesNotExist = add(1005, "data does not exist")
	// 重复值
	ErrDuplicate = add(1006, "duplicate value")
	// 请求参数有误
	ErrWrongParameter = add(1007, "wrong parameter")

	// 登录失败
	ErrLoginFailed = add(2001, "login failed")
	// 用户不存在
	ErrUserNotFound = add(2002, "account not found")
	// 手机号不正确
	ErrPhoneNumber = add(2003, "phone number is incorrect")
	// IP地址不正确
	ErrIPAddress = add(2004, "ip address is incorrect")
)
