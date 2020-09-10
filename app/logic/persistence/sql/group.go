package sql

import (
	"context"
	"outgoing/app/logic/persistence"
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

	getGroupsSQL = `
SELECT
	g.created_at,
	g.name,
	g.gid,
	g.introduction,
	g.owner,
	g.member_count
FROM
	group_member gm
JOIN 
	public.group g
ON 
	g.ID = gm.group_id
AND
	g.activated = true
WHERE
	gm.user_id = $1
ORDER BY
	created_at DESC;`

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

func (p *groupPersister) Create(_ context.Context, in *persistence.GroupCreate) (*persistence.Group, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Unix()
	if err = tx.Exec(insertGroupSQL, 1, in.GroupID, now, in.ClientID, in.Name, in.GID, in.Introduction, in.Owner, 0, true); err != nil {
		return nil, err
	}

	if err = increaseClientGroupCount(tx, in.ClientID, 1); err != nil {
		return nil, err
	}

	if err = tx.Exec(insertGroupMemberSQL, 1, now, in.GroupID, in.Owner); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &persistence.Group{
		CreatedAt:    now,
		Name:         in.Name,
		GID:          in.GID,
		Introduction: in.Introduction,
		Owner:        in.Owner,
		MemberCount:  1,
	}, nil
}

func (p *groupPersister) GetGroups(_ context.Context, userID int64) ([]*persistence.Group, error) {
	rows, err := p.db.Query(getGroupsSQL, userID)
	if err != nil {
		return nil, err
	}

	var groups []*persistence.Group
	for rows.Next() {
		var group persistence.Group
		if err := rows.Scan(&group.CreatedAt, &group.Name, &group.GID, &group.Introduction, &group.Owner, &group.MemberCount); err != nil {
			return nil, err
		}

		groups = append(groups, &group)
	}

	return groups, nil
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

func (p *groupPersister) GetMembers(_ context.Context, clientID string, groupID int64) ([]int64, error) {
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

	var memberIDs []int64
	for rows.Next() {
		var memberID int64
		if err := rows.Scan(&memberID); err != nil {
			return nil, err
		}

		memberIDs = append(memberIDs, memberID)
	}

	return memberIDs, nil
}
