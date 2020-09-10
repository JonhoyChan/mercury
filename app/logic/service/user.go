package service

import (
	"context"
	"outgoing/app/service/api"
	"outgoing/app/service/persistence"
	"outgoing/x/types"
)

func (s *Service) CreateUser(ctx context.Context, req *api.CreateUserReq) (string, error) {
	s.log.Info("[CreateUser] request is received")

	clientID := s.MustGetContextClient(ctx)
	id := s.idGen.Get()
	in := &persistence.UserCreate{
		ClientID: clientID,
		UserID:   s.idGen.DecodeID(id),
		Name:     req.Name,
		UID:      id.UID(),
	}
	if err := s.persister.User().Create(ctx, in); err != nil {
		s.log.Error("[CreateUser] failed to create user", "client_id", clientID, "name", req.Name, "error", err)
		return "", err
	}

	return id.UID(), nil
}

func (s *Service) UpdateActivated(ctx context.Context, uid string, activated bool) error {
	s.log.Info("[UpdateActivated] request is received")

	if err := s.persister.User().UpdateActivated(ctx, s.DecodeID(types.ParseUID(uid)), activated); err != nil {
		s.log.Error("[UpdateActivated] failed to update user activated", "uid", uid, "activated", activated, "error", err)
		return err
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, uid string) error {
	s.log.Info("[DeleteUser] request is received")

	if err := s.persister.User().Delete(ctx, s.DecodeID(types.ParseUID(uid))); err != nil {
		s.log.Error("[DeleteUser] failed to delete client", "uid", uid, "error", err)
		return err
	}

	return nil
}

func (s *Service) AddFriend(ctx context.Context, uid, friendUID string) error {
	s.log.Info("[DeleteUser] request is received")

	clientID := s.MustGetContextClient(ctx)
	in := &persistence.UserFriend{
		ClientID:     clientID,
		UserID:       s.idGen.DecodeID(types.ParseUID(uid)),
		FriendUserID: s.idGen.DecodeID(types.ParseUID(friendUID)),
	}
	if err := s.persister.User().AddFriend(ctx, in); err != nil {
		s.log.Error("[DeleteUser] failed to delete client", "uid", uid, "error", err)
		return err
	}

	return nil
}

func (s *Service) DeleteFriend(ctx context.Context, uid, friendUID string) error {
	s.log.Info("[DeleteUser] request is received")

	clientID := s.MustGetContextClient(ctx)
	in := &persistence.UserFriend{
		ClientID:     clientID,
		UserID:       s.idGen.DecodeID(types.ParseUID(uid)),
		FriendUserID: s.idGen.DecodeID(types.ParseUID(friendUID)),
	}
	if err := s.persister.User().DeleteFriend(ctx, in); err != nil {
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
