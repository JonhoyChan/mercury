package cache

import (
	"mercury/x/database/redis"
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	c := &Cache{
		client: client,
	}
	return c
}

// Check redis connection
func (c *Cache) Ping() error {
	return c.client.Ping().Err()
}

// Close redis connection
func (c *Cache) Close() error {
	return c.client.Close()
}
