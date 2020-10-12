package sql

import (
	"context"
	"mercury/app/logic/persistence"
	"mercury/x/database/sqlx"
	"mercury/x/ecode"
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
        uid,
		activated
    )
VALUES
    ($1, $2, $2, $3, $4, $5, $6);
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

	getFriendsSQL = `
SELECT
	friend_user_id
FROM
	friend
WHERE
	user_id = $1
ORDER BY
	created_at DESC;
`

	deleteFriendSQL = `
DELETE FROM
    friend
WHERE
    user_id = $1
AND
	friend_user_id = $2;
`
)

func (p *userPersister) CheckActivated(_ context.Context, clientID, uid string) (bool, error) {
	var isExist int
	err := p.db.QueryRow("SELECT 1 FROM public.user WHERE client_id = $1 AND uid = $2 AND activated = $3 limit 1;", clientID, uid, true).
		Scan(&isExist)
	if err != nil {
		if sqlx.IsErrNoRows(err) {
			return false, nil
		}
		return false, err
	}

	return isExist == 1, nil
}

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
		err = ecode.ErrDataAlreadyExists
		return err
	}

	now := time.Now().Unix()
	if err = tx.Exec(insertUserSQL, 1, in.UserID, now, in.ClientID, in.Name, in.UID, true); err != nil {
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

func (p *userPersister) UpdateActivated(_ context.Context, id int64, activated bool) error {
	return p.db.Exec("UPDATE public.user SET activated = $1 WHERE id = $2;", 1, activated, id)
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
		return ecode.ErrDataAlreadyExists
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

func (p *userPersister) GetFriends(_ context.Context, userID int64) ([]int64, error) {
	rows, err := p.db.Query(getFriendsSQL, userID)
	if err != nil {
		return nil, err
	}

	var friendIDs []int64
	for rows.Next() {
		var friendID int64
		if err := rows.Scan(&friendID); err != nil {
			return nil, err
		}

		friendIDs = append(friendIDs, friendID)
	}

	return friendIDs, nil
}

func (p *userPersister) DeleteFriend(_ context.Context, in *persistence.UserFriend) error {
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
	if err = tx.QueryRow("SELECT 1 FROM public.user WHERE client_id = $1 AND id = $2 limit 1;", in.ClientID, in.FriendUserID).Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if err = tx.Exec(deleteFriendSQL, 1, in.UserID, in.FriendUserID); err != nil {
		return err
	}
	if err = tx.Exec(deleteFriendSQL, 1, in.FriendUserID, in.UserID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
