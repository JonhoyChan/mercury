package sql

import (
	"outgoing/app/service/chat/persistence"
	"outgoing/x/database/sqlx"
)

type Persister struct {
	db      *sqlx.DB
	client  *clientPersister
	user    *userPersister
	message *messagePersister
}

func NewPersister(db *sqlx.DB) *Persister {
	return &Persister{
		db: db,
		client: &clientPersister{
			db: db,
		},
		user: &userPersister{
			db: db,
		},
		message: &messagePersister{
			db: db,
		},
	}
}

func (p *Persister) Ping() error {
	return p.db.Ping()
}

func (p *Persister) Close() error {
	return p.db.Close()
}

func (p *Persister) Client() persistence.ClientPersister {
	return p.client
}

func (p *Persister) User() persistence.UserPersister {
	return p.user
}

func (p *Persister) Message() persistence.MessagePersister {
	return p.message
}
