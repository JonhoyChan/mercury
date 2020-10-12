package redis

import (
	"mercury/x/config"

	"github.com/go-redis/redis/v7"
)

var (
	RedisNil = redis.Nil
)

type Client struct {
	*redis.Client
}

func NewClient(c config.RedisProvider) (*Client, error) {
	rc := c.Redis()
	options := &redis.Options{
		Addr:        rc.Address,
		Username:    rc.Username,
		Password:    rc.Password,
		DB:          rc.DB,
		IdleTimeout: rc.IdleTimeout,
	}

	client := redis.NewClient(options)

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &Client{client}, nil
}
