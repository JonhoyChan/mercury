package sql

import (
	"context"
	"outgoing/app/service/chat/persistence"
	"outgoing/x/database/sqlx"
	"outgoing/x/ecode"
	"time"
)

type userPersister struct {
	db *sqlx.DB
}

const (
	insertUserSQL = `
INSERT INTO
    public.user (
		id,
        created_at,
        updated_at,
		client_id,
        name,
        uid
    )
VALUES
    ($1, $2, $2, $3, $4, $5);
`

	isFriendExistSQL = `
SELECT
    1
FROM
    friend
WHERE
    user_id = $1
AND
	friend_user_id = $2
limit
    1;
`

	insertFriendSQL = `
INSERT INTO
    friend (
        created_at,
        updated_at,
        user_id,
        friend_user_id
    )
VALUES
    ($1, $1, $2, $3),
    ($1, $1, $3, $2);
`

	deleteFriendSQL = `
DELETE FROM
    friend
WHERE
    user_id = $1
AND
	friend_user_id = $2
);
DELETE FROM
    friend
WHERE
    user_id = $2
AND
	friend_user_id = $1
);
`
)

func (p *userPersister) Create(_ context.Context, in *persistence.UserCreate) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var isExist int
	if err = tx.QueryRow("SELECT 1 FROM public.user WHERE client_id = $1 AND name = $2 limit 1;", in.ClientID, in.Name).
		Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if isExist == 1 {
		err = ecode.ErrDataAlreadyExist
		return err
	}

	now := time.Now().Unix()
	if err = tx.Exec(insertUserSQL, 1, in.UserID, now, in.ClientID, in.Name, in.UserID); err != nil {
		return err
	}

	if err = increaseClientUserCount(tx, in.ClientID, 1); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *userPersister) Delete(_ context.Context, id int64) error {
	return p.db.Exec("DELETE FROM public.user WHERE id = $1;", 1, id)
}

func (p *userPersister) AddFriend(_ context.Context, in *persistence.UserFriend) error {
	var isExist int
	if err := p.db.QueryRow(isFriendExistSQL, in.UserID, in.FriendUserID).Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if isExist == 1 {
		return ecode.ErrDataAlreadyExist
	}

	if err := p.db.QueryRow("SELECT 1 FROM public.user WHERE client_id = $1 AND id = $2 limit 1;", in.ClientID, in.FriendUserID).Scan(&isExist); err != nil && err != sqlx.ErrNoRows {
		return err
	}

	if isExist == 0 {
		return ecode.ErrDataDoesNotExist
	}

	now := time.Now().Unix()
	if err := p.db.Exec(insertFriendSQL, 2, now, in.UserID, in.FriendUserID); err != nil {
		return err
	}

	return nil
}

func (p *userPersister) DeleteFriend(_ context.Context, in *persistence.UserFriend) error {
	var isExist int
	if err := p.db.QueryRow("SELECT 1 FROM public.user WHERE client_id = $1 AND id = $2 limit 1;", in.ClientID, in.FriendUserID).Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if err := p.db.Exec(deleteFriendSQL, 2, in.ClientID, in.FriendUserID); err != nil {
		return err
	}

	return nil
}
