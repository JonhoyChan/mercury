package service

import (
	"context"
	chatApi "mercury/app/logic/api"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/types"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
)

type Servicer interface {
	SessionStore() SessionStore
	Close()
}

type Service struct {
	chatService  chatApi.ChatService
	log          log.Logger
	sessionStore SessionStore
}

func NewService(l log.Logger) (*Service, error) {
	opts := []client.Option{
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Service{
		chatService:  chatApi.NewChatService("mercury.logic", c),
		log:          l,
		sessionStore: NewSessionStore(),
	}, nil
}

func (s *Service) SessionStore() SessionStore {
	return s.sessionStore
}

func (s *Service) Close() {
	s.sessionStore.Shutdown()
}

func (s *Service) connect(ctx context.Context, token, sid, serverID string) (string, types.ID, error) {
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

func (s *Service) readMessage(ctx context.Context, uid, topic string, sequence int64) error {
	_, err := s.chatService.ReadMessage(ctx, &chatApi.ReadMessageReq{
		UID:      uid,
		Topic:    topic,
		Sequence: sequence,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) keypress(ctx context.Context, uid, topic string) error {
	_, err := s.chatService.Keypress(ctx, &chatApi.KeypressReq{
		UID:   uid,
		Topic: topic,
	})
	if err != nil {
		return err
	}
	return nil
}
