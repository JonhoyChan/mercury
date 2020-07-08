package cache

import (
	"time"
	"outgoing/x"
	"outgoing/x/ecode"
)

var (
	authTokenKey = "string/auth/token/%v"
)

// 获取用户Token
func (c *Cache) GetAuthToken(uid string) string {
	key := x.Sprintf(authTokenKey, uid)
	return c.client.Get(key).String()
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
