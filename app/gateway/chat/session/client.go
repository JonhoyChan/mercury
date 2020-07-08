package session

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	uApi "outgoing/app/service/account/api"
	aApi "outgoing/app/service/auth/api"
	cApi "outgoing/app/service/chat/api"
	"outgoing/x/ecode"
	"outgoing/x/types"
)

var globalClient = NewClient()

type Client struct {
	accountService uApi.AccountService
	authService    aApi.AuthService
	chatService    cApi.ChatService
}

func NewClient() *Client {
	opts := []client.Option{
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Client{
		accountService: uApi.NewAccountService("account.srv", c),
		authService:    aApi.NewAuthService("auth.srv", c),
		chatService:    cApi.NewChatService("auth.srv", c),
	}
}

func (c *Client) authenticate(ctx context.Context, token, sid, serverID string) (types.Uid, types.AuthLevel, error) {
	resp, err := c.authService.Authenticate(ctx, &aApi.AuthenticateReq{
		HandlerType: aApi.HandlerType_HandlerTypeJWT,
		Token:       token,
	})
	if err != nil {
		return 0, 0, err
	}

	_, err = c.chatService.Connect(ctx, &cApi.ConnectReq{
		UID:      resp.Record.Uid,
		SID:      sid,
		ServerID: serverID,
	})
	if err != nil {
		return 0, 0, err
	}

	return types.ParseUid(resp.Record.Uid), types.AuthLevel(resp.Record.Level), nil
}
