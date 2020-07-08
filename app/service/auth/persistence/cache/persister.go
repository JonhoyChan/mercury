package cache

import (
	"outgoing/x/database/redis"
	"outgoing/x/log"
)

type Cache struct {
	client *redis.Client
	log    log.Logger
}

func NewCache(client *redis.Client, log log.Logger) *Cache {
	c := &Cache{
		client: client,
		log:    log.New("cache", "redis"),
	}
	return c
}

func (c *Cache) Ping() error {
	return c.client.Ping().Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
