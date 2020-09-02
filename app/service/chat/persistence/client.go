package persistence

import (
	"time"
)

type Client struct {
	ID          string
	Name        string
	TokenSecret []byte
	TokenExpire time.Duration
}

type ClientCreate struct {
	ID          string
	Name        string
	TokenSecret string
	Credential  string
	TokenExpire int64
}

type ClientUpdate struct {
	ID          string
	Name        *string
	TokenSecret *string
	TokenExpire *int64
}
