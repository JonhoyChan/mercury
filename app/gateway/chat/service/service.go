package service

import (
	"context"
	chatApi "outgoing/app/service/chat/api"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"outgoing/x/types"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
)

type Service struct {
	SessionStore SessionStore
	log          log.Logger
	chatService  chatApi.ChatService
}

func NewService(log log.Logger) *Service {
	opts := []client.Option{
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Service{
		SessionStore: NewSessionStore(),
		log:          log,
		chatService:  chatApi.NewChatService("service.chat.logic", c),
	}
}

func (s *Service) authenticate(ctx context.Context, token, sid, serverID string) (types.Uid, types.AuthLevel, error) {
	resp, err := s.chatService.Authenticate(ctx, &chatApi.AuthenticateReq{
		Token: token,
	})
	if err != nil {
		return 0, 0, err
	}

	_, err = s.chatService.Connect(ctx, &chatApi.ConnectReq{
		UID:      resp.Record.UID,
		SID:      sid,
		ServerID: serverID,
	})
	if err != nil {
		return 0, 0, err
	}

	return types.ParseUserUID(resp.Record.UID), types.AuthLevel(resp.Record.Level), nil
}

func (s *Service) heartbeat(ctx context.Context, uid, sid, serverID string) (err error) {
	_, err = s.chatService.Heartbeat(ctx, &chatApi.HeartbeatReq{
		UID:      uid,
		SID:      sid,
		ServerID: serverID,
	})
	return
}
