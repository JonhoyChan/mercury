package sql

import (
	"context"
	"outgoing/app/service/chat/persistence"
	"outgoing/x"
	"outgoing/x/database/sqlx"
	"outgoing/x/ecode"
	"outgoing/x/types"
	"time"
)

type messagePersister struct {
	db *sqlx.DB
}

const (
	insertMessageSQL = `
INSERT INTO
    message (
        created_at,
        updated_at,
		topic,
		sequence,
		message_type,
		sender,
        receiver,
		content_type,
		body,
		mentions,
		status
    )
VALUES
    ($1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;
`
)

func (p *messagePersister) Add(_ context.Context, message *persistence.Message) error {
	var isExist int
	if err := p.db.QueryRow("SELECT 1 FROM message WHERE topic = $1 AND sequence = $2 limit 1;", message.Topic, message.Sequence).
		Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if isExist == 1 {
		return ecode.ErrDataAlreadyExists
	}

	message.CreatedAt = time.Now().Unix()
	// use the QueryRow instead of ExecX, because LastInsertId is not supported by this driver
	err := p.db.QueryRow(insertMessageSQL, message.CreatedAt, message.Topic, message.Sequence, message.MessageType,
		message.Sender, message.Receiver, message.ContentType, message.Body, x.Join(message.Mentions, ","),
		types.MessageStatusNormal).
		Scan(&message.ID)
	if err != nil {
		return err
	}

	return nil
}

func (p *messagePersister) GetTopicLastSequence(_ context.Context, topic string) (int64, error) {
	var sequence int64
	if err := p.db.QueryRow("SELECT sequence FROM message WHERE topic = $1 ORDER BY sequence DESC LIMIT 1;", topic).
		Scan(&sequence); sqlx.IsErrNoRows(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return sequence, nil
}
