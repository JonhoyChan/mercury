package config

import "time"

type Redis struct {
	Address     string        `json:"address"`
	Username    string        `json:"username"`
	Password    string        `json:"password"`
	DB          int           `json:"db"`
	IdleTimeout time.Duration `json:"idle_timeout"`
}

func DefaultRedis() *Redis {
	return &Redis{
		Address:     "localhost:6379",
		Username:    "",
		Password:    "",
		DB:          0,
		IdleTimeout: 120 * time.Second,
	}
}
