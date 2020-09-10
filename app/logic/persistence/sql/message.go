package sql

import (
	"context"
	"outgoing/app/logic/persistence"
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
		status,
		mentions
    )
VALUES
    ($1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;
`

	getMessagesBySequenceSQL = `
SELECT
	id,
    created_at,
	topic,
	sequence,
	message_type,
	sender,
	receiver,
	content_type,
	body,
	status,
	mentions
FROM
    message
WHERE
    topic = $1
AND
	sequence = $2
LIMIT 1;
`

	getMessagesByLastSequenceSQL = `
SELECT
	id,
    created_at,
	topic,
	sequence,
	message_type,
	sender,
	receiver,
	content_type,
	body,
	status,
	mentions
FROM
    message
WHERE
    topic = $1
AND
	sequence > $2
ORDER BY
    sequence DESC;
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
		message.Sender, message.Receiver, message.ContentType, message.Body, types.MessageStatusNormal,
		x.Join(message.Mentions, ",")).
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

func (p *messagePersister) GetTopicMessageBySequence(_ context.Context, topic string, sequence int64) (*persistence.Message, error) {
	var (
		message                  persistence.Message
		messageType, contentType uint8
		body, mentions           string
	)
	if err := p.db.QueryRow(getMessagesBySequenceSQL, topic, sequence).Scan(&message.ID, &message.CreatedAt,
		&message.Topic, &message.Sequence, &messageType, &message.Sender, &message.Receiver,
		&contentType, &body, &message.Status, &mentions); err != nil {
		if sqlx.IsErrNoRows(err) {
			return nil, ecode.ErrDataDoesNotExist
		}
		return nil, err
	}

	message.MessageType = types.MessageType(messageType)
	message.ContentType = types.ContentType(contentType)
	message.Body = []byte(body)
	if mentions != "" {
		message.Mentions = x.SplitInt64(mentions, ",")
	}

	return &message, nil
}

func (p *messagePersister) GetTopicMessagesByLastSequence(_ context.Context, topic string, sequence int64) ([]*persistence.Message, int64, error) {
	var count int64
	if err := p.db.QueryRow("SELECT count(*) FROM message WHERE topic = $1 AND sequence > $2", topic, sequence).
		Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := p.db.Query(getMessagesByLastSequenceSQL, topic, sequence)
	if err != nil {
		return nil, 0, err
	}

	var messages []*persistence.Message
	for rows.Next() {
		var (
			message                  persistence.Message
			messageType, contentType uint8
			body, mentions           string
		)
		if err := rows.Scan(&message.ID, &message.CreatedAt, &message.Topic, &message.Sequence, &messageType,
			&message.Sender, &message.Receiver, &contentType, &body, &message.Status, &mentions); err != nil {
			return nil, 0, err
		}

		message.MessageType = types.MessageType(messageType)
		message.ContentType = types.ContentType(contentType)
		message.Body = []byte(body)
		if mentions != "" {
			message.Mentions = x.SplitInt64(mentions, ",")
		}

		messages = append(messages, &message)
	}

	return messages, count, nil
}
