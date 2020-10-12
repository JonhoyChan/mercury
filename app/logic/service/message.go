package service

import (
	"context"
	jsoniter "github.com/json-iterator/go"
	"mercury/app/logic/api"
	"mercury/app/logic/persistence"
	"mercury/x"
	"mercury/x/database/redis"
	"mercury/x/ecode"
	"mercury/x/types"
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
			s.log.Error("[PushMessage] receiver not activated", "uid", req.Receiver)
			return 0, 0, ecode.ErrUserNotActivated
		}

		uids = append(uids, req.Receiver)
		topic = sender.P2PName(receiver)

		go s.cache.SetUsersTopic(uids, topic)
	case api.MessageTypeGroup:
		gid := types.ParseGID(req.Receiver)
		receiver = gid
		// Get a list of all member IDs in the group
		members, err := s.persister.Group().GetMembers(ctx, req.ClientID, s.DecodeID(gid))
		if err != nil {
			s.log.Error("[PushMessage] failed to get group members", "gid", req.Receiver, "error", err)
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
		// Get the next sequence of the topic
		message.Sequence, err = s.nextSequence(ctx, topic)
		if err != nil {
			s.log.Error("[PushMessage] failed to get sequence", "topic", topic, "error", err)
			return err
		}

		if err = s.persister.Message().Add(ctx, message); err != nil {
			s.log.Error("[PushMessage] failed to add message", "sequence", message.Sequence, "topic", topic, "error", err)
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
				go s.send(types.OperationPush, m, req.SID, uids...)
				if len(req.Mentions) > 0 {
					n := &types.Notification{
						Topic: topic,
						What:  types.WhatTypeMentioned,
					}
					go s.send(types.OperationNotification, n, "", req.Mentions...)
				}
				go s.cache.SetUserTopicLastSequence(sender.UID(), message.Topic, message.Sequence)

				go s.InvokeMessageListener(req.ClientID, m)
				return message.ID, message.Sequence, nil
			}

			// If the error is ErrDataAlreadyExists, it means that the sequence already exists,
			// try to get the next sequence again,
			// otherwise return this error
			if !ecode.EqualError(ecode.ErrDataAlreadyExists, err) {
				return 0, 0, err
			}
		}
	}

	return 0, 0, ecode.ErrInternalServer
}

func (s *Service) PullMessage(ctx context.Context, req *api.PullMessageReq) ([]*api.TopicMessages, error) {
	topicsLastSequence, err := s.cache.GetUserTopicsLastSequence(req.UID)
	if err != nil {
		return nil, err
	}

	if len(topicsLastSequence) == 0 {
		topics, err := s.cache.GetUserTopics(req.UID)
		if err != nil {
			return nil, err
		}

		if len(topics) > 0 {
			topicsLastSequence = make(map[string]int64)
			for _, topic := range topics {
				topicsLastSequence[topic] = 0
			}
		}
	}

	var topicMessages []*api.TopicMessages
	for topic, sequence := range topicsLastSequence {
		messages, count, err := s.persister.Message().GetTopicMessagesByLastSequence(ctx, topic, sequence)
		if err != nil {
			return nil, err
		}

		tm := &api.TopicMessages{
			Topic: topic,
			Count: count,
		}

		for _, message := range messages {
			mentions := make([]string, 0)
			for _, mention := range message.Mentions {
				mentions = append(mentions, s.EncodeID(mention).UID())
			}
			tm.Messages = append(tm.Messages, &api.Message{
				ID:          message.ID,
				CreatedAt:   message.CreatedAt,
				MessageType: message.MessageType.String(),
				Sender:      s.EncodeID(message.Sender).UID(),
				Receiver:    s.EncodeID(message.Receiver).UID(),
				Topic:       message.Topic,
				Sequence:    message.Sequence,
				ContentType: message.ContentType.String(),
				Body:        message.Body,
				Mentions:    mentions,
			})
		}

		topicMessages = append(topicMessages, tm)
	}

	return topicMessages, nil
}

func (s *Service) ReadMessage(ctx context.Context, req *api.ReadMessageReq) error {
	message, err := s.persister.Message().GetTopicMessageBySequence(ctx, req.Topic, req.Sequence)
	if err != nil {
		return err
	}

	err = s.cache.SetUserTopicLastSequence(req.UID, message.Topic, message.Sequence)
	if err != nil {
		return err
	}

	if s.EncodeID(message.Receiver).Compare(types.ParseUID(req.UID)) == 0 {
		n := &types.Notification{
			Topic:     message.Topic,
			What:      types.WhatTypeRead,
			Sequence:  message.Sequence,
			MessageID: message.ID,
		}
		sender := s.EncodeID(message.Sender).UID()
		go s.send(types.OperationNotification, n, "", sender)
	}

	return nil
}

func (s *Service) Keypress(ctx context.Context, req *api.KeypressReq) error {
	from := types.ParseUID(req.UID)

	u1, u2, err := types.ParseP2P(req.Topic)
	if err != nil {
		return err
	}

	var to types.ID
	if u1.Compare(from) == 0 {
		to = u2
	} else if u2.Compare(from) == 0 {
		to = u1
	}

	n := &types.Notification{
		Topic: req.Topic,
		What:  types.WhatTypeKeypress,
	}
	go s.send(types.OperationNotification, n, "", to.UID())

	return nil
}

func (s *Service) send(op types.Operation, v interface{}, skipSID string, uids ...string) {
	sessions, _, err := s.cache.GetSessions(uids...)
	if err != nil {
		s.log.Warn("[Send] failed to get sessions", "error", err)
		return
	}

	if len(sessions) > 0 {
		servers := make(map[string][]string)
		for sid, serverID := range sessions {
			if sid == "" || serverID == "" {
				s.log.Warn("[Send] sid or serverID is empty", "sid", sid, "serverID", serverID, "error", err)
				continue
			}
			if sid != skipSID {
				servers[serverID] = append(servers[serverID], sid)
			}
		}

		data, err := jsoniter.Marshal(v)
		if err != nil {
			s.log.Warn("[Send] failed to marshal", "error", err)
			return
		}

		topic := s.config.PushMessageTopic()
		for serverID, sids := range servers {
			if err := s.invoke(topic, &api.PushMessage{
				Operation: int32(op),
				ServerID:  serverID,
				SIDs:      sids,
				Data:      data,
			}); err != nil {
				s.log.Warn("[Send] failed to invoke", "serverID", serverID, "error", err)
			}
		}
	}
}
