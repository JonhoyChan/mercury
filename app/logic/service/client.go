package service

import (
	"context"
	"github.com/google/uuid"
	"mercury/app/logic/api"
	"mercury/app/logic/persistence"
	"mercury/x"
	"mercury/x/ecode"
)

func (s *Service) getClient(ctx context.Context, clientID string) (client *persistence.Client, err error) {
	client, err = s.cache.GetClient(clientID)
	if err != nil {
		client, err = s.persister.Client().GetClient(ctx, clientID)
		if err != nil {
			return
		}

		go s.cache.SetClient(clientID, client)
	}

	return
}

func (s *Service) GetClient(ctx context.Context) (*api.Client, error) {
	id := MustClientIDFromContext(ctx)
	client, err := s.getClient(ctx, id)
	if err != nil {
		s.log.Error("[GetClient] failed to get client", "error", err)
		return nil, err
	}

	return &api.Client{
		ID:          client.ID,
		CreatedAt:   client.CreatedAt,
		UpdatedAt:   client.UpdatedAt,
		Name:        client.Name,
		TokenSecret: client.TokenSecret,
		TokenExpire: int64(client.TokenExpire.Seconds()),
		UserCount:   client.UserCount,
		GroupCount:  client.GroupCount,
	}, nil
}

func (s *Service) CreateClient(ctx context.Context, req *api.CreateClientReq) (string, string, error) {
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
	id := MustClientIDFromContext(ctx)
	in := &persistence.ClientUpdate{
		ID: id,
	}
	if req.Name != nil {
		in.Name = &req.Name.Value
	}
	if req.TokenSecret != nil {
		in.TokenSecret = &req.TokenSecret.Value
	}
	if req.TokenExpire != nil {
		in.TokenExpire = &req.TokenExpire.Value
	}
	if err := s.persister.Client().Update(ctx, in); err != nil {
		s.log.Error("[UpdateClient] failed to update client", "client_id", id, "error", err)
		return err
	}

	go func() {
		client, err := s.persister.Client().GetClient(ctx, id)
		if err != nil {
			s.log.Error("[UpdateClient] failed to get client", "client_id", id, "error", err)
			return
		}

		err = s.cache.SetClient(id, client)
		if err != nil {
			s.log.Error("[UpdateClient] failed to set client to cache", "client_id", id, "error", err)
		}
	}()

	return nil
}

func (s *Service) DeleteClient(ctx context.Context) error {
	id := MustClientIDFromContext(ctx)
	if err := s.persister.Client().Delete(ctx, id); err != nil {
		s.log.Error("[DeleteClient] failed to delete client", "client_id", id, "error", err)
		return err
	}

	go func() {
		err := s.cache.DeleteClient(id)
		if err != nil {
			s.log.Error("[DeleteClient] failed to delete client from cache", "client_id", id, "error", err)
		}
	}()

	return nil
}

func (s *Service) GenerateToken(ctx context.Context, req *api.GenerateTokenReq) (string, string, error) {
	credential, err := s.persister.Client().GetClientCredential(ctx, req.ClientID)
	if err != nil {
		s.log.Error("[GenerateToken] failed to get client credential", "client_id", req.ClientID, "error", err)
		return "", "", err
	}

	if err = s.hash.Compare([]byte(credential), []byte(req.ClientSecret)); err != nil {
		s.log.Error("[GenerateToken] failed to compare", "error", err)
		return "", "", err
	}

	token, lifetime, err := s.token.GenerateToken(req.ClientID)
	if err != nil {
		s.log.Error("[GenerateToken] failed to generate token", "client_id", req.ClientID, "error", err)
		return "", "", err
	}

	return token, lifetime, nil
}

func (s *Service) Listen(ctx context.Context, token string, stream api.ChatClientAdmin_ListenStream) error {
	var clientID string
	_, err := s.token.Authenticate(token, &clientID)
	if err != nil {
		s.log.Error("[Listen] failed to authenticating the token", "error", err)
		return ecode.ErrInvalidToken
	}

	return s.listen(clientID, stream)
}
