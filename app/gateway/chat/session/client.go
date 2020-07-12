package session

import (
	"context"
	accountApi "outgoing/app/service/account/api"
	authApi "outgoing/app/service/auth/api"
	chatApi "outgoing/app/service/chat/api"
	"outgoing/x/ecode"
	"outgoing/x/types"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
)

var globalClient = NewClient()

type Client struct {
	accountService accountApi.AccountService
	authService    authApi.AuthService
	chatService    chatApi.ChatService
}

func NewClient() *Client {
	opts := []client.Option{
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Client{
		accountService: accountApi.NewAccountService("service.account", c),
		authService:    authApi.NewAuthService("service.auth", c),
		chatService:    chatApi.NewChatService("service.chat.logic", c),
	}
}

func (c *Client) authenticate(ctx context.Context, token, sid, serverID string) (types.Uid, types.AuthLevel, error) {
	resp, err := c.authService.Authenticate(ctx, &authApi.AuthenticateReq{
		HandlerType: authApi.HandlerType_HandlerTypeJWT,
		Token:       token,
	})
	if err != nil {
		return 0, 0, err
	}

	_, err = c.chatService.Connect(ctx, &chatApi.ConnectReq{
		UID:      resp.Record.Uid,
		SID:      sid,
		ServerID: serverID,
	})
	if err != nil {
		return 0, 0, err
	}

	return types.ParseUserUID(resp.Record.Uid), types.AuthLevel(resp.Record.Level), nil
}

func (c *Client) heartbeat(ctx context.Context, uid, sid, serverID string) (err error) {
	_, err = c.chatService.Heartbeat(ctx, &chatApi.HeartbeatReq{
		UID:      uid,
		SID:      sid,
		ServerID: serverID,
	})
	return
}

//func (c *Client) getUser(ctx context.Context, uid types.Uid) (*User, error) {
//	resp, err := c.accountService.GetUser(ctx, &accountApi.GetUserReq{UID: uid.UID()})
//	if err != nil {
//		return nil, err
//	}
//
//	user := &User{
//		Uid:       types.ParseUid(resp.User.Uid),
//		CreatedAt: time.Unix(resp.User.CreatedAt, 0),
//		UpdatedAt: time.Unix(resp.User.UpdatedAt, 0),
//		NickName:  resp.User.NickName,
//		Avatar:    resp.User.Avatar,
//		Gender:    resp.User.Gender,
//		Bio:       resp.User.Bio,
//		Birthday:  resp.User.Birthday,
//		Mobile:    resp.User.Mobile,
//		State:     resp.User.State,
//	}
//
//	return user, nil
//}
//
//func (c *Client) getUsers(ctx context.Context, uids ...types.Uid) ([]*User, error) {
//	req := &accountApi.GetUsersReq{}
//	for _, uid := range uids {
//		req.UIDs = append(req.UIDs, uid.UID())
//	}
//	resp, err := c.accountService.GetUsers(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//
//	var users []*User
//	for i := range resp.Users {
//		user := resp.Users[i]
//		users = append(users, &User{
//			Uid:       types.ParseUid(user.Uid),
//			CreatedAt: time.Unix(user.CreatedAt, 0),
//			UpdatedAt: time.Unix(user.UpdatedAt, 0),
//			NickName:  user.NickName,
//			Avatar:    user.Avatar,
//			Gender:    user.Gender,
//			Bio:       user.Bio,
//			Birthday:  user.Birthday,
//			Mobile:    user.Mobile,
//			State:     user.State,
//		})
//	}
//
//	return users, nil
//}
