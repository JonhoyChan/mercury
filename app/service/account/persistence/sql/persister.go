package sql

import (
	"outgoing/app/service/account/persistence"
	"outgoing/app/service/account/persistence/cache"
	"outgoing/x/database/sqlx"
	"outgoing/x/log"
	"outgoing/x/password"
)

type persister struct {
	db   *sqlx.DB
	user persistence.UserPersister
}

func NewPersister(db *sqlx.DB, c *cache.Cache, hasher password.Hasher, log log.Logger) *persister {
	l := log.New("persister", "sql")
	p := &persister{
		db: db,
		user: &userPersister{
			db:     db,
			hasher: hasher,
			log:    l,
			c:      c,
		},
	}
	return p
}

func (p *persister) Ping() error {
	return p.db.Ping()
}

func (p *persister) Close() error {
	return p.db.Close()
}

func (p *persister) User() persistence.UserPersister {
	return p.user
}
