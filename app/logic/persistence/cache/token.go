package cache

import (
	"mercury/x"
	"mercury/x/ecode"
	"time"
)

const (
	tokenKey = "token:%v"
)

func (c *Cache) GetClientID(token string) string {
	key := x.Sprintf(tokenKey, token)
	clientID, _ := c.client.Get(key).Result()
	return clientID
}

func (c *Cache) SetClientID(token, clientID string, lifetime time.Duration) error {
	if token == "" {
		return ecode.NewError("token is missing")
	} else if clientID == "" {
		return ecode.NewError("client ID is missing")
	} else if lifetime <= 0 {
		return ecode.NewError("invalid lifetime")
	}
	key := x.Sprintf(tokenKey, token)
	if err := c.client.Set(key, clientID, lifetime).Err(); err != nil {
		return err
	}

	return nil
}
