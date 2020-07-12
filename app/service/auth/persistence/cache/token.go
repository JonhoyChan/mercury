package cache

import (
	"outgoing/x"
	"outgoing/x/ecode"
	"time"
)

var (
	authTokenKey = "string/auth/token/%v"
)

// 获取用户Token
func (c *Cache) GetAuthToken(uid string) string {
	key := x.Sprintf(authTokenKey, uid)
	token, err := c.client.Get(key).Result()
	if err != nil {
		return ""
	}
	return token
}

// 保存用户Token
func (c *Cache) SetAuthToken(uid, token string, lifetime time.Duration) error {
	if uid == "" {
		return ecode.NewError("the uid is missing")
	}
	if token == "" {
		return ecode.NewError("the token is missing")
	}
	if lifetime <= 0 {
		return ecode.NewError("invalid expiration value")
	}
	key := x.Sprintf(authTokenKey, uid)
	if err := c.client.Set(key, token, lifetime).Err(); err != nil {
		return err
	}

	return nil
}
