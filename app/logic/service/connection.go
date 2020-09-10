package service

import (
	"context"
	"outgoing/app/logic/api"
	"outgoing/x/ecode"
)

func (s *Service) Connect(ctx context.Context, req *api.ConnectReq) (string, string, error) {
	s.log.Info("[Connect] request is received")

	clientID := s.cache.GetClientID(req.JWTToken)
	if clientID == "" {
		return "", "", ecode.ErrInvalidToken
	}
	client, err := s.getClient(ctx, clientID)
	if err != nil {
		s.log.Error("[Connect] failed to get client", "client_id", clientID, "error", err)
		return "", "", err
	}

	var uid string
	_, err = s.jwt.Authenticate(req.JWTToken, client.Name, client.TokenSecret, &uid)
	if err != nil {
		s.log.Error("[Connect] failed to authenticating the jwt token", "uid", uid, "error", err)
		return "", "", err
	}

	if err := s.cache.AddMapping(uid, req.SID, req.ServerID); err != nil {
		s.log.Error("[Connect] failed to add mapping", "uid", uid, "error", err)
		return "", "", err
	}

	return clientID, uid, nil
}

func (s *Service) Disconnect(ctx context.Context, req *api.DisconnectReq) error {
	s.log.Info("[Disconnect] request is received")

	if err := s.cache.DeleteMapping(req.UID, req.SID); err != nil {
		s.log.Error("[Disconnect] failed to delete mapping", "uid", req.UID, "error", err)
		return err
	}

	return nil
}

func (s *Service) Heartbeat(ctx context.Context, req *api.HeartbeatReq) error {
	s.log.Info("[Heartbeat] request is received")

	expired, err := s.cache.ExpireMapping(req.UID, req.SID)
	if err != nil {
		s.log.Error("[Heartbeat] failed to expire mapping", "uid", req.UID, "error", err)
		return err
	}
	if !expired {
		if err := s.cache.AddMapping(req.UID, req.SID, req.ServerID); err != nil {
			s.log.Error("[Heartbeat] failed to add mapping", "uid", req.UID, "error", err)
			return err
		}
	}

	return nil
}
