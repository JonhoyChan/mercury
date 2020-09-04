package service

import (
	"context"
	"github.com/micro/go-micro/v2/broker"
	cApi "outgoing/app/gateway/api"
	"outgoing/app/service/api"
	"outgoing/x/ecode"
	"outgoing/x/log"
)

func (s *Service) subscribePushMessage(e broker.Event) error {
	log.Info("subscribe", "topic", e.Topic())

	if e.Message() == nil {
		return ecode.NewError("message can not be nil")
	}

	bodyBytes := e.Message().Body

	pm := new(api.PushMessage)
	if err := pm.Unmarshal(bodyBytes); err != nil {
		return err
	}
	if err := s.push(context.Background(), pm); err != nil {
		return err
	}

	return nil
}

func (s *Service) push(ctx context.Context, pm *api.PushMessage) error {
	var err error
	switch pm.Type {
	case api.PushMessageTypeDefault:
		err = s.pushDefault(ctx, pm.ServerID, pm.SIDs, pm.Data)
	default:
		err = ecode.NewError("can not match push type")
	}
	return err
}

func (s *Service) pushDefault(ctx context.Context, serverID string, sids []string, data []byte) error {
	if comet, ok := s.cometServers[serverID]; ok {
		comet.Push(&cApi.PushMessageReq{
			SIDs: sids,
			Data: data,
		})
	}
	return nil
}
