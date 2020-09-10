package service

import (
	"context"
	"encoding/json"
	"outgoing/app/service/api"
	"outgoing/app/service/persistence"
	"outgoing/x"
	"outgoing/x/database/redis"
	"outgoing/x/ecode"
	"outgoing/x/types"
)

func (s *Service) nextSequence(ctx context.Context, topic string) (int64, error) {
	sequence, err := s.cache.IncrTopicSequence(topic)
	switch err {
	case redis.RedisNil:
		sequence, err = s.persister.Message().GetTopicLastSequence(ctx, topic)
		if err != nil {
			return 0, err
		}
		sequence++
		go s.cache.SetTopicSequence(topic, sequence, 0)
		return sequence, nil
	case nil:
		return sequence, nil
	default:
		return 0, err
	}
}

const addMessageRetries = 3

func (s *Service) PushMessage(ctx context.Context, req *api.PushMessageReq) (int64, int64, error) {
	sender := types.ParseUID(req.Sender)

	check, _ := s.persister.User().CheckActivated(ctx, req.ClientID, req.Sender)
	if !check {
		s.log.Error("[SendMessage] sender not activated", "uid", req.Sender)
		return 0, 0, ecode.ErrUserNotActivated
	}

	uids := []string{req.Sender}
	var (
		receiver types.ID
		topic    string
		mentions []int64
	)
	switch req.MessageType {
	case api.MessageTypeSingle:
		receiver = types.ParseUID(req.Receiver)
		check, _ = s.persister.User().CheckActivated(ctx, req.ClientID, req.Receiver)
		if !check {
			s.log.Error("[SendMessage] receiver not activated", "uid", req.Receiver)
			return 0, 0, ecode.ErrUserNotActivated
		}

		uids = append(uids, req.Receiver)
		topic = sender.P2PName(receiver)
	case api.MessageTypeGroup:
		gid := types.ParseGID(req.Receiver)
		receiver = gid
		// Get a list of all member IDs in the group
		members, err := s.persister.Group().Members(ctx, req.ClientID, s.DecodeID(gid))
		if err != nil {
			s.log.Error("[SendMessage] failed to get group members", "gid", req.Receiver, "error", err)
			return 0, 0, err
		}

		// Check if the sender is in the group
		if !x.IsInSlice(members, s.DecodeID(sender)) {
			return 0, 0, ecode.NewError("The user is not join the group")
		}

		for _, member := range members {
			uids = append(uids, s.EncodeID(member).UID())
		}
		for _, mention := range req.Mentions {
			id := s.DecodeID(types.ParseUID(mention))
			if !x.IsInSlice(members, id) {
				return 0, 0, ecode.ErrInternalServer.ResetMessage("The mentioned user is not join the group")
			}
			mentions = append(mentions, id)
		}
		topic = req.Receiver
	}

	message := &persistence.Message{
		Topic:       topic,
		MessageType: types.MessageType(req.MessageType),
		Sender:      s.DecodeID(sender),
		Receiver:    s.DecodeID(receiver),
		ContentType: types.ContentType(req.ContentType),
		Body:        req.Body,
		Mentions:    mentions,
	}
	add := func() error {
		var err error
		message.Sequence, err = s.nextSequence(ctx, topic)
		if err != nil {
			s.log.Error("[SendMessage] failed to get sequence", "topic", topic, "error", err)
			return err
		}

		if err = s.persister.Message().Add(ctx, message); err != nil {
			s.log.Error("[SendMessage] failed to add message", "sequence", message.Sequence, "topic", topic, "error", err)
			return err
		}

		return nil
	}

	ch := make(chan error, addMessageRetries)
	for i := 0; i < addMessageRetries; i++ {
		go func() {
			ch <- add()
		}()

		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		case err := <-ch:
			if err == nil {
				m := &types.Message{
					ID:          message.ID,
					CreatedAt:   message.CreatedAt,
					MessageType: types.MessageType(req.MessageType),
					Sender:      sender.UID(),
					Receiver:    req.Receiver,
					Topic:       topic,
					Sequence:    message.Sequence,
					ContentType: types.ContentType(req.ContentType),
					Body:        req.Body,
					Mentions:    req.Mentions,
				}

				sessions, _, err := s.cache.GetSessions(uids...)
				if err != nil {
					s.log.Error("[SendMessage] failed to get sessions", "error", err)
					return 0, 0, err
				}

				if len(sessions) > 0 {
					servers := make(map[string][]string)
					for sid, serverID := range sessions {
						if sid == "" || serverID == "" {
							s.log.Warn("[SendMessage] sid or serverID is empty", "sid", sid, "serverID", serverID, "error", err)
							continue
						}
						if sid != req.SID {
							servers[serverID] = append(servers[serverID], sid)
						}
					}

					data, err := json.Marshal(m)
					if err != nil {
						return 0, 0, err
					}
					for serverID, sids := range servers {
						if err := s.invoke(s.config.PushMessageTopic(), &api.PushMessage{
							ServerID: serverID,
							SIDs:     sids,
							Data:     data,
						}); err != nil {
							s.log.Warn("[SendMessage] failed to invoke", "serverID", serverID, "error", err)
						}
					}
				}

				go s.InvokeMessageListener(req.ClientID, m)
				return message.ID, message.Sequence, nil
			}

			if !ecode.EqualError(ecode.ErrDataAlreadyExists, err) {
				return 0, 0, err
			}
		}
	}

	return 0, 0, ecode.ErrInternalServer
}
