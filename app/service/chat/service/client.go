package service

import (
	"context"
	"github.com/google/uuid"
	"outgoing/app/service/chat/api"
	"outgoing/app/service/chat/persistence"
	"outgoing/x"
)

func (s *Service) GenerateToken(ctx context.Context, req *api.GenerateTokenReq) (string, string, error) {
	s.log.Info("[GenerateToken] request is received")

	credential, err := s.persister.Client().GetClientCredential(ctx, req.ClientID)
	if err != nil {
		s.log.Error("[GenerateToken] failed to get client credential", "client_id", req.ClientID, "error", err)
		return "", "", err
	}

	if err = s.hash.Compare([]byte(credential), []byte(req.ClientSecret)); err != nil {
		s.log.Error("[GenerateToken] failed to compare", "error", err)
		return "", "", err
	}

	token, lifetime, err := s.multiHandler[api.HandlerTypeDefault].GenerateToken(req.ClientID)
	if err != nil {
		s.log.Error("[GenerateToken] failed to generate token", "client_id", req.ClientID, "error", err)
		return "", "", err
	}

	return token, lifetime, nil
}

func (s *Service) CreateClient(ctx context.Context, req *api.CreateClientReq) (string, string, error) {
	s.log.Info("[CreateClient] request is received")

	secret, err := x.GenerateSecret(26)
	if err != nil {
		s.log.Error("[CreateClient] failed to generate secret", "error", err)
		return "", "", err
	}
	credential, err := s.hash.Hash(secret)
	if err != nil {
		s.log.Error("[CreateClient] failed to create a hash from secret", "error", err)
		return "", "", err
	}
	id := uuid.New().String()
	in := &persistence.ClientCreate{
		ID:          id,
		Name:        req.Name,
		TokenSecret: req.TokenSecret,
		Credential:  string(credential),
		TokenExpire: req.TokenExpire,
	}
	if err := s.persister.Client().Create(ctx, in); err != nil {
		s.log.Error("[CreateClient] failed to create client", "client_name", req.Name, "error", err)
		return "", "", err
	}

	return id, string(secret), nil
}

func (s *Service) UpdateClient(ctx context.Context, req *api.UpdateClientReq) error {
	s.log.Info("[UpdateClient] request is received")

	id := s.MustGetClientID(ctx)
	in := &persistence.ClientUpdate{
		ID:          id,
		Name:        &req.ClientName,
		TokenSecret: &req.TokenSecret,
		TokenExpire: &req.TokenExpire,
	}
	if err := s.persister.Client().Update(ctx, in); err != nil {
		s.log.Error("[CreateClient] failed to update client", "client_id", id, "error", err)
		return err
	}

	return nil
}

func (s *Service) DeleteClient(ctx context.Context) error {
	s.log.Info("[DeleteClient] request is received")

	id := s.MustGetClientID(ctx)
	if err := s.persister.Client().Delete(ctx, id); err != nil {
		s.log.Error("[DeleteClient] failed to delete client", "client_id", id, "error", err)
		return err
	}

	return nil
}
