package service

import (
	"context"
	"outgoing/app/service/api"
	"outgoing/app/service/persistence"
	"outgoing/x/types"
)

func (s *Service) CreateGroup(ctx context.Context, req *api.CreateGroupReq) (string, error) {
	s.log.Info("[CreateGroup] request is received")

	clientID := s.MustGetContextClient(ctx)
	uid := s.idGen.Get()
	in := &persistence.GroupCreate{
		ClientID:     clientID,
		GroupID:      s.idGen.DecodeID(uid),
		Name:         req.Name,
		GID:          uid.GID(),
		Introduction: req.Introduction,
		Owner:        s.idGen.DecodeID(types.ParseUID(req.Owner)),
	}
	if err := s.persister.Group().Create(ctx, in); err != nil {
		s.log.Error("[CreateGroup] failed to create group", "client_id", clientID, "name", req.Name, "owner", req.Owner, "error", err)
		return "", err
	}

	return uid.GID(), nil
}

func (s *Service) AddMember(ctx context.Context, req *api.AddMemberReq) error {
	s.log.Info("[AddMember] request is received")

	clientID := s.MustGetContextClient(ctx)
	in := &persistence.GroupMember{
		ClientID: clientID,
		GroupID:  s.idGen.DecodeID(types.ParseGID(req.GID)),
		UserID:   s.idGen.DecodeID(types.ParseUID(req.UID)),
	}
	if err := s.persister.Group().AddMember(ctx, in); err != nil {
		s.log.Error("[AddMember] failed to add member to group", "client_id", clientID, "gid", req.GID, "uid", req.UID, "error", err)
		return err
	}

	// TODO send the notification to other group members.

	return nil
}
