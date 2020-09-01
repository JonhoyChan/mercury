package cache

import (
	"muhub/x"
	"muhub/x/ecode"
	"time"
)

var (
	authTokenKey = "authToken:%v"
)

// 获取用户Token
func (c *Cache) GetAuthToken(uid string) string {
	key := x.Sprintf(authTokenKey, uid)
	token, _ := c.client.Get(key).Result()
	return token
}

func (c *Cache) SetAuthToken(uid, token string, lifetime time.Duration) error {
	if uid == "" {
		return ecode.NewError("the uid is missing")
	} else if token == "" {
		return ecode.NewError("the token is missing")
	} else if lifetime <= 0 {
		return ecode.NewError("invalid lifetime")
	}
	key := x.Sprintf(authTokenKey, uid)
	if err := c.client.Set(key, token, lifetime).Err(); err != nil {
		return err
	}

	return nil
}
