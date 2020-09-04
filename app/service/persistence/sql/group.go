package sql

import (
	"context"
	"outgoing/app/service/persistence"
	"outgoing/x/database/sqlx"
	"outgoing/x/ecode"
	"time"
)

type groupPersister struct {
	db *sqlx.DB
}

const (
	isGroupExistSQL = `
SELECT
    1
FROM
    public.group
WHERE
    client_id = $1
AND
	id = $2
limit
    1;
`

	insertGroupSQL = `
INSERT INTO
    public.group (
		id,
        created_at,
        updated_at,
		client_id,
        name,
        gid,
		introduction,
		owner,
		type,
		activated,
		member_count
    )
VALUES
    ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, 1);
`
	isGroupMemberExistSQL = `
SELECT
    1
FROM
    public.group_member
WHERE
    group_id = $1
AND
	user_id = $2
limit
    1;
`

	insertGroupMemberSQL = `
INSERT INTO
    public.group_member (
        created_at,
        updated_at,
        group_id,
        user_id
    )
VALUES
    ($1, $1, $2, $3);
`
)

func (p *groupPersister) Create(_ context.Context, in *persistence.GroupCreate) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Unix()
	if err = tx.Exec(insertGroupSQL, 1, in.GroupID, now, in.ClientID, in.Name, in.GID, in.Introduction, in.Owner, 0, true); err != nil {
		return err
	}

	if err = increaseClientGroupCount(tx, in.ClientID, 1); err != nil {
		return err
	}

	if err = tx.Exec(insertGroupMemberSQL, 1, now, in.GroupID, in.Owner); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func increaseGroupMemberCount(tx *sqlx.Tx, id int64, count int64) error {
	return tx.Exec("UPDATE public.group SET member_count = member_count + $1 WHERE id = $2;", 1, count, id)
}

func (p *groupPersister) AddMember(_ context.Context, in *persistence.GroupMember) error {
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
	if err = tx.QueryRow(isGroupExistSQL, in.ClientID, in.GroupID).Scan(&isExist); err != nil && err != sqlx.ErrNoRows {
		return err
	}

	if isExist == 0 {
		return ecode.ErrDataDoesNotExist
	}

	isExist = 0
	if err = tx.QueryRow(isGroupMemberExistSQL, in.GroupID, in.UserID).Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if isExist == 1 {
		return ecode.ErrDataAlreadyExists
	}

	now := time.Now().Unix()
	if err = tx.Exec(insertGroupMemberSQL, 1, now, in.GroupID, in.UserID); err != nil {
		return err
	}

	if err = increaseGroupMemberCount(tx, in.GroupID, 1); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *groupPersister) CheckMember(_ context.Context, groupID int64, userID int64) (bool, error) {
	var isExist int
	if err := p.db.QueryRow(isGroupMemberExistSQL, groupID, userID).Scan(&isExist); err != nil {
		if sqlx.IsErrNoRows(err) {
			return false, nil
		}
		return false, err
	}

	return isExist == 1, nil
}

func (p *groupPersister) Members(_ context.Context, clientID string, groupID int64) ([]int64, error) {
	var isExist int
	if err := p.db.QueryRow(isGroupExistSQL, clientID, groupID).Scan(&isExist); err != nil && err != sqlx.ErrNoRows {
		return nil, err
	}

	if isExist == 0 {
		return nil, ecode.ErrDataDoesNotExist
	}

	rows, err := p.db.Query("SELECT user_id FROM group_member WHERE group_id = $1 ORDER BY created_at DESC;", groupID)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (p *groupPersister) Groups(_ context.Context, userID int64) ([]int64, error) {
	// FIXME
	rows, err := p.db.Query("SELECT group_id FROM group_member WHERE user_id = $1 ORDER BY created_at DESC;", userID)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
