package persistence

import (
	"time"
)

type Client struct {
	ID          string
	CreatedAt   int64
	UpdatedAt   int64
	Name        string
	TokenSecret []byte
	TokenExpire time.Duration
	UserCount   int64
	GroupCount  int64
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
