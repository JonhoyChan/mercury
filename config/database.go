package config

import "time"

type Database struct {
	Driver      string        `json:"driver"`
	DSN         string        `json:"dsn"`
	Active      int           `json:"active"`
	Idle        int           `json:"idle"`
	IdleTimeout time.Duration `json:"idle_timeout"`
}

func DefaultDatabase() *Database {
	return &Database{
		Driver:      "postgres",
		DSN:         "postgresql://root@localhost:26257/mercury?sslmode=disable",
		Active:      10,
		Idle:        5,
		IdleTimeout: 4 * time.Hour,
	}
}
