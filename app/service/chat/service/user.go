package service

import (
	"context"
	"outgoing/app/service/chat/api"
	"outgoing/app/service/chat/persistence"
	"outgoing/x/types"
)

func (s *Service) CreateUser(ctx context.Context, req *api.CreateUserReq) (string, error) {
	s.log.Info("[CreateUser] request is received")

	clientID := s.MustGetContextClient(ctx)
	uid := s.uidGen.Get()
	in := &persistence.UserCreate{
		ClientID: clientID,
		UserID:   s.uidGen.DecodeUid(uid),
		Name:     req.Name,
		UID:      uid.UID(),
	}
	if err := s.persister.User().Create(ctx, in); err != nil {
		s.log.Error("[CreateUser] failed to create user", "client_id", clientID, "name", req.Name, "error", err)
		return "", err
	}

	return uid.UID(), nil
}

func (s *Service) UpdateActivated(ctx context.Context, uid string, activated bool) error {
	s.log.Info("[UpdateActivated] request is received")

	if err := s.persister.User().UpdateActivated(ctx, s.DecodeUid(types.ParseUID(uid)), activated); err != nil {
		s.log.Error("[UpdateActivated] failed to update user activated", "uid", uid, "activated", activated, "error", err)
		return err
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, uid string) error {
	s.log.Info("[DeleteUser] request is received")

	if err := s.persister.User().Delete(ctx, s.DecodeUid(types.ParseUID(uid))); err != nil {
		s.log.Error("[DeleteUser] failed to delete client", "uid", uid, "error", err)
		return err
	}

	return nil
}

func (s *Service) GenerateUserToken(ctx context.Context, uid string) (string, string, error) {
	s.log.Info("[GenerateUserToken] request is received")

	clientID := s.MustGetContextClient(ctx)
	client, err := s.getClient(ctx, clientID)
	if err != nil {
		s.log.Error("[GenerateUserToken] failed to get client", "client_id", clientID, "error", err)
		return "", "", err
	}

	token, lifetime, err := s.jwt.GenerateToken(client.Name, client.TokenSecret, client.TokenExpire, uid)
	if err != nil {
		s.log.Error("[GenerateUserToken] failed to generate token", "uid", uid, "error", err)
		return "", "", err
	}

	go s.cache.SetClientID(token, clientID, client.TokenExpire)

	return token, lifetime, nil
}
