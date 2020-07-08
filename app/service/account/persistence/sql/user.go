package sql

import (
	"context"
	"fmt"
	"time"
	"outgoing/app/service/main/account/model"
	"outgoing/app/service/main/account/persistence"
	"outgoing/x"
	"outgoing/x/database/sqlx"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"outgoing/x/password"
	"outgoing/x/types"
)

type userPersister struct {
	db     *sqlx.DB
	log    log.Logger
	hasher password.Hasher
	c      persistence.Cacher
}

const (
	insertUserSQL = `
INSERT INTO
    public.account (
		id,
        created_at,
        updated_at,
        vid,
        nick_name,
        avatar,
        gender,
        mobile,
        state
    )
VALUES
    ($1, $2, $2, $3, $4, $5, 0, $6, 0);
`

	insertUserAuthSQL = `
INSERT INTO
    public.user_auth (
        created_at,
        updated_at,
        user_id,
        identity_type,
        identifier,
        credential,
        last_at
    )
VALUES
    ($1, $1, $2, 1, $3, $4, $1),
    ($1, $1, $2, 2, $5, $4, $1);
`

	insertUserRegisterLogSQL = `
INSERT INTO
    public.user_register_log (created_at, user_id, method, ip)
VALUES
    ($1, $2, $3, $4);
`

	insertUserLoginLogSQL = `
INSERT INTO
    public.user_login_log (created_at, user_id, version, method, command, device_id, ip)
VALUES
    ($1, $2, $3, $4, $5, $6, $7);
`

	getCredentialViaMobileSQL = `
SELECT
    u.id,
    u.vid,
	u.state,
    ua.credential
FROM
    public.account u
    JOIN public.user_auth ua ON u.id = ua.user_id
WHERE
    u.mobile ~ $1
    AND ua.identity_type = $2
limit
    1;
`
)

func (p *userPersister) createUserRegisterLog(registerAt, userID int64, method int8, ip string) {
	err := p.db.Exec(insertUserRegisterLogSQL, 1, registerAt, userID, method, ip)
	if err != nil {
		p.log.Warn("failed to create account register log", "error", err)
	}
}

func (p *userPersister) createUserLoginLog(createdAt, userID int64, version, deviceID, ip string, method, command int8) {
	err := p.db.Exec(insertUserLoginLogSQL, 1, createdAt, userID, version, method, command, deviceID, ip)
	if err != nil {
		p.log.Warn("failed to create account login log", "error", err)
	}
}

func (p *userPersister) Register(_ context.Context, id int64, uid types.Uid, mobile, avatar, ip string) (string, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var isExist int
	if err = tx.QueryRow("SELECT 1 FROM public.account WHERE mobile = $1 limit 1;", mobile).Scan(&isExist); err != nil && err != sqlx.ErrNoRows {
		fmt.Println(err.Error())
		return "", err
	}

	if isExist == 1 {
		err = ecode.Wrap(ecode.ErrDataAlreadyExist, "该手机号已被注册")
		return "", err
	}

	now := time.Now().Unix()
	// 生成默认vid，后续可由用户自行修改，只能改一次还是限制一段时间内改一次待定
	vid := uid.PrefixId("vid")

	if err := tx.Exec(insertUserSQL, 1, id, now, vid, uid.String32(), avatar, mobile); err != nil {
		return "", err
	}

	pwd, _ := p.hasher.Generate("")
	// 将用户手机与vid保存到用户授权表中，用户可使用两者其中一种+密码的方式进行登录
	err = tx.Exec(insertUserAuthSQL, 2,
		now, id, mobile, pwd, vid,
	)
	if err != nil {
		return "", err
	}

	// TODO 判断地理位置是否为空，如果不为空保存用户的地理位置信息

	// TODO 保存用户注册的设备信息

	if err := tx.Commit(); err != nil {
		return "", err
	}

	go p.createUserRegisterLog(now, id, 1, ip)

	return vid, nil
}

func (p *userPersister) cleanFailedCount(vid string) {
	if err := p.c.CleanFailedCount(vid); err != nil {
		p.log.Warn("failed to clean account login failed count", "error", err)
	}
}

func (p *userPersister) LoginViaPassword(_ context.Context, uc *model.UserCredential, pwd, version, deviceID, ip string) error {
	ok, ttl, err := p.c.ContinueLogin(uc.VID)
	if err != nil {
		return err
	}

	if !ok {
		return ecode.Wrap(ecode.ErrLoginFailed, x.Sprintf("暂时无法登录，请等待%d分钟后再试", int(ttl.Minutes())))
	}

	now := time.Now().Unix()
	if err := p.hasher.Compare(pwd, uc.Credential); err != nil {
		remain, err := p.c.IncreaseFailedCount(uc.VID)
		if err != nil {
			return err
		}
		go p.createUserLoginLog(now, uc.ID, version, deviceID, ip, 1, 2)
		return ecode.Wrap(ecode.ErrLoginFailed, x.Sprintf("账号或密码错误，还剩%d次重试机会", remain))
	}

	go p.cleanFailedCount(uc.VID)
	go p.createUserLoginLog(now, uc.ID, version, deviceID, ip, 1, 1)
	return nil
}

func (p *userPersister) LoginViaCaptcha(_ context.Context, captcha, version, deviceID, ip string) error {
	return nil
}

func (p *userPersister) GetCredentialViaMobile(mobile string) (*model.UserCredential, error) {
	var (
		id              int64
		vid, credential string
		state           uint8
	)
	err := p.db.QueryRow(getCredentialViaMobileSQL, mobile+`$`, 1).
		Scan(&id, &vid, &state, &credential)
	if err != nil {
		if err == sqlx.ErrNoRows {
			return nil, ecode.Wrap(ecode.ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}

	if state != 0 {
		switch state {
		case 1: // 用户被暂停
			// todo something
		case 2: // 用户已注销
			return nil, ecode.Wrap(ecode.ErrUserNotFound, "用户已注销")
		}
	}

	userCredential := &model.UserCredential{
		ID:         id,
		VID:        vid,
		Credential: credential,
	}

	return userCredential, nil
}
