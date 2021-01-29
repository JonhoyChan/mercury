package service

import (
	"context"
	"mercury/app/logic/api"
	"mercury/app/logic/persistence"
	"mercury/x/types"
)

func (s *Service) CreateGroup(ctx context.Context, req *api.CreateGroupReq) (*api.Group, error) {
	clientID := MustClientIDFromContext(ctx)
	id := s.idGen.Get()
	in := &persistence.GroupCreate{
		ClientID:     clientID,
		GroupID:      s.idGen.DecodeID(id),
		Name:         req.Name,
		GID:          id.GID(),
		Introduction: req.Introduction,
		Owner:        s.idGen.DecodeID(types.ParseUID(req.Owner)),
	}
	group, err := s.persister.Group().Create(ctx, in)
	if err != nil {
		s.log.Error("[CreateGroup] failed to create group", "client_id", clientID, "name", req.Name, "owner", req.Owner, "error", err)
		return nil, err
	}

	go s.cache.SetUsersTopic([]string{req.Owner}, group.GID)

	return &api.Group{
		CreatedAt:    group.CreatedAt,
		Name:         group.Name,
		GID:          group.GID,
		Introduction: group.Introduction,
		Owner:        req.Owner,
		MemberCount:  group.MemberCount,
	}, nil
}

func (s *Service) GetGroups(ctx context.Context, uid string) ([]*api.Group, error) {
	clientID := MustClientIDFromContext(ctx)
	groups, err := s.persister.Group().GetGroups(ctx, s.DecodeID(types.ParseUID(uid)))
	if err != nil {
		s.log.Error("[GetGroups] failed to get groups", "client_id", clientID, "uid", uid, "error", err)
		return nil, err
	}

	var result []*api.Group
	for i := 0; i < len(groups); i++ {
		group := groups[i]
		result = append(result, &api.Group{
			CreatedAt:    group.CreatedAt,
			Name:         group.Name,
			GID:          group.GID,
			Introduction: group.Introduction,
			Owner:        s.EncodeID(group.Owner).UID(),
			MemberCount:  group.MemberCount,
		})
	}

	return result, nil
}

func (s *Service) AddMember(ctx context.Context, req *api.AddMemberReq) error {
	clientID := MustClientIDFromContext(ctx)
	in := &persistence.GroupMember{
		ClientID: clientID,
		GroupID:  s.idGen.DecodeID(types.ParseGID(req.GID)),
		UserID:   s.idGen.DecodeID(types.ParseUID(req.UID)),
	}
	if err := s.persister.Group().AddMember(ctx, in); err != nil {
		s.log.Error("[AddMember] failed to add member to group", "client_id", clientID, "gid", req.GID, "uid", req.UID, "error", err)
		return err
	}

	go s.cache.SetUsersTopic([]string{req.UID}, req.GID)

	// TODO send the notification to other group members.

	return nil
}

func (s *Service) GetMembers(ctx context.Context, gid string) ([]string, error) {
	clientID := MustClientIDFromContext(ctx)
	memberIDs, err := s.persister.Group().GetMembers(ctx, clientID, s.DecodeID(types.ParseGID(gid)))
	if err != nil {
		s.log.Error("[GetMembers] failed to get group members", "gid", gid, "error", err)
		return nil, err
	}

	var result []string
	for i := 0; i < len(memberIDs); i++ {
		result = append(result, s.EncodeID(memberIDs[i]).UID())
	}

	return result, nil
}
