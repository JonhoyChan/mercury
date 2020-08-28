package service

import (
	"context"
	"outgoing/app/service/chat/api"
	"outgoing/app/service/chat/persistence"
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

func (s *Service) UpdateActivated(ctx context.Context, activated bool) error {
	s.log.Info("[UpdateActivated] request is received")

	u := s.MustGetContextUser(ctx)
	if err := s.persister.User().UpdateActivated(ctx, u.ID, activated); err != nil {
		s.log.Error("[UpdateActivated] failed to update user activated", "uid", u.UID, "activated", activated, "error", err)
		return err
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context) error {
	s.log.Info("[DeleteUser] request is received")

	u := s.MustGetContextUser(ctx)
	if err := s.persister.User().Delete(ctx, u.ID); err != nil {
		s.log.Error("[DeleteUser] failed to delete client", "uid", u.UID, "error", err)
		return err
	}

	return nil
}

func (s *Service) GenerateUserToken(ctx context.Context) (string, string, error) {
	s.log.Info("[GenerateUserToken] request is received")

	clientID := s.MustGetContextClient(ctx)
	client, err := s.persister.Client().GetClient(ctx, clientID)
	if err != nil {
		s.log.Error("[GenerateUserToken] failed to get client", "client_id", clientID, "error", err)
		return "", "", err
	}

	u := s.MustGetContextUser(ctx)
	token, lifetime, err := s.jwt.GenerateToken(client.Name, []byte(client.TokenSecret), client.TokenExpire, u.UID)
	if err != nil {
		s.log.Error("[GenerateUserToken] failed to generate token", "uid", u.UID, "error", err)
		return "", "", err
	}

	return token, lifetime, nil
}
