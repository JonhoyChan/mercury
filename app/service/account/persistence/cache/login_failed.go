package cache

import (
	"outgoing/x"
	"outgoing/x/database/redis"
	"time"
)

var (
	loginFailedKey = "string/login/failed/%v"
)

const (
	// The limit of the account login error
	loginFailedLimit = 5
	// The expiration of the account login error
	loginFailedExpiration = 10 * 60
)

// Determine whether the user can continue to login
func (c *Cache) ContinueLogin(uid string) (bool, time.Duration, error) {
	key := x.Sprintf(loginFailedKey, uid)
	failedCount, err := c.client.Get(key).Int64()
	if err != nil {
		if err == redis.RedisNil {
			return true, 0, nil
		}
		c.log.Error("failed to get", "key", key, "error", err)
		return false, 0, err
	}

	if failedCount >= loginFailedLimit {
		ttl, err := c.client.TTL(key).Result()
		if err != nil {
			c.log.Error("failed to ttl", "key", key, "error", err)
			return false, 0, err
		}

		return false, ttl, nil
	}

	return true, 0, nil
}

// Increase the number of login failures
func (c *Cache) IncreaseFailedNumber(uid string) (int64, error) {
	key := x.Sprintf(loginFailedKey, uid)
	failedNumber, err := c.client.Get(key).Int64()
	if err != nil && err != redis.RedisNil {
		c.log.Error("failed to get", "key", key, "error", err)
		return 0, err
	}

	if failedNumber < loginFailedLimit {
		failedNumber += 1
	}

	expiration := loginFailedExpiration * time.Second

	if err := c.client.Set(key, failedNumber, expiration).Err(); err != nil {
		c.log.Error("failed to set", "key", key, "value", failedNumber, "error", err)
		return 0, err
	}

	return loginFailedLimit - failedNumber, nil
}

// Clean the number of login failures
func (c *Cache) CleanFailedNumber(uid string) error {
	key := x.Sprintf(loginFailedKey, uid)
	if err := c.client.Del(key).Err(); err != nil {
		c.log.Error("failed to del", "key", key, "error", err)
		return err
	}

	return nil
}
