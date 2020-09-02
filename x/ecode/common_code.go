package ecode

var (
	OK           = add(200, "ok")
	ResetContent = add(205, "reset content")

	// Bad request
	ErrBadRequest = add(400, "bad request")
	// Unauthorized
	ErrUnauthorized = add(401, "unauthorized")
	// Forbidden
	ErrForbidden = add(403, "forbidden")
	// Not found
	ErrNotFound = add(404, "not found")
	// Method not allowed
	ErrMethodNotAllowed = add(405, "method not allowed")
	// Request timeout
	ErrRequestTimeout = add(408, "request timeout")
	// Locked
	ErrLocked = add(423, "locked")
	// Too many requests
	ErrTooManyRequests = add(429, "too many requests")
	// Internal server
	ErrInternalServer = add(500, "internal server")

	// Invalid token
	ErrInvalidToken = add(1001, "invalid token")
	// TToken expired
	ErrTokenExpired = add(1002, "token expired")
	// Malformed
	ErrMalformed = add(1003, "malformed")
	// Data already exists
	ErrDataAlreadyExists = add(1004, "data already exists")
	// Data does not exist
	ErrDataDoesNotExist = add(1005, "data does not exist")
	// Wrong parameter
	ErrWrongParameter = add(1006, "wrong parameter")

	// User not activated
	ErrUserNotActivated = add(2001, "user not activated")
)
