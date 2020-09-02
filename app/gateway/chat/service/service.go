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
		chatService:  chatApi.NewChatService("mercury.logic", c),
	}
}

func (s *Service) connect(ctx context.Context, token, sid, serverID string) (string, types.Uid, error) {
	resp, err := s.chatService.Connect(ctx, &chatApi.ConnectReq{
		JWTToken: token,
		SID:      sid,
		ServerID: serverID,
	})
	if err != nil {
		return "", 0, err
	}

	return resp.ClientID, types.ParseUID(resp.UID), nil
}

func (s *Service) heartbeat(ctx context.Context, uid, sid, serverID string) error {
	_, err := s.chatService.Heartbeat(ctx, &chatApi.HeartbeatReq{
		UID:      uid,
		SID:      sid,
		ServerID: serverID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) pushMessage(ctx context.Context, req *chatApi.PushMessageReq) (int64, int64, error) {
	resp, err := s.chatService.PushMessage(ctx, req)
	if err != nil {
		return 0, 0, err
	}
	return resp.MessageId, resp.Sequence, nil
}
