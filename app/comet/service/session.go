package service

import (
	"context"
	chatApi "mercury/app/logic/api"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/types"
	"time"

	"mercury/x/websocket"
)

// Wire transport
const (
	NONE = iota
	WEBSOCKET
)

// Wait time before abandoning the outbound send operation.
// Timeout is rather long to make sure it's longer than Linux preemption time:
// https://elixir.bootlin.com/linux/latest/source/kernel/sched/fair.c#L38
const sendTimeout = time.Millisecond * 7

// Maximum number of queued messages before session is considered stale and dropped.
const sendQueueLimit = 128

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = time.Second * 55

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// TODO change to configurable
var (
	// Maximum message size allowed from peer.
	MaxMessageSize int64 = 1 << 19 // 512K
	// minSupportedVersion is the minimum supported version
	MinSupportedVersion = "0.1.0-dev"
)

var minSupportedVersionValue = x.ParseVersion(MinSupportedVersion)

// Session represents a single WS connection or a long polling session. A user may have multiple
// sessions.
type Session struct {
	ctx context.Context
	// Session ID.
	sid string
	// Server ID.
	serverID string
	// protocol - NONE (unset), WEBSOCKET.
	proto int
	// Websocket. Set only for websocket sessions.
	ws *websocket.Connection
	// IP address of the client.
	remoteAddress string
	// User agent identifying client software
	userAgent string
	// Protocol version of the client: ((major & 0xff) << 8) | (minor & 0xff).
	version int
	// Device ID of the client.
	deviceID string
	// Platform: web, ios, android.
	platform string
	// Human language of the client.
	language string
	// ID of the client to which the current user belongs
	clientID string
	// ID of the current user or 0.
	id types.ID
	// Time when the session received any packer from client.
	lastAction time.Time
	// Outbound messages, buffered.
	// The content must be serialized in format suitable for the session.
	send chan []byte
	// Channel for shutting down the session, buffer 1.
	// Content in the same format as for 'send'.
	stop chan []byte
	// Service
	srv *Service
}

func (s *Session) readLoop() {
	defer func() {
		s.ws.Close()
		s.srv.sessionStore.Delete(s)
	}()

	s.ws.SetReadLimit(MaxMessageSize)
	s.ws.SetReadDeadline(time.Now().Add(pongWait))
	s.ws.SetPongHandler(func(string) error {
		s.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		raw, err := s.ws.ReadMessage()
		if err != nil {
			log.Error("[Websocket] failed to read message", log.Ctx{"error": err, "sid": s.sid})
			return
		}

		if raw == nil {
			return
		}

		s.dispatchRaw(raw)
	}
}

func (s *Session) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		// Break readLoop.
		s.ws.Close()
	}()

	for {
		select {
		case msg, ok := <-s.send:
			if !ok {
				// Channel closed.
				return
			}
			if len(s.send) > sendQueueLimit {
				log.Warn("[Websocket] outbound queue limit exceeded", log.Ctx{"sid": s.sid})
				return
			}
			if err := s.ws.WriteBinaryMessage(msg); err != nil {
				log.Error("[Websocket] failed to write binary message", log.Ctx{"error": err, "sid": s.sid})
				return
			}
		case msg := <-s.stop:
			// Shutdown requested, don't care if the message is delivered
			if msg != nil {
				_ = s.ws.WriteTextMessage(msg)
			}
			return
		case <-ticker.C:
			if err := s.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("[Websocket] failed to write ping message", log.Ctx{"error": err, "sid": s.sid})
				return
			}
		}
	}
}

// Message received, convert bytes to ClientComMessage and dispatch
func (s *Session) dispatchRaw(raw []byte) {
	var p Protocol
	if err := p.Unmarshal(raw); err != nil {
		s.queueOut(&p, ErrMalformed("", 0))
		return
	}

	if !p.Validate() {
		s.queueOut(&p, ErrMalformed("", 0))
		return
	}

	s.dispatch(&p)
}

type handlerFunc func(message *ServerMessage) []byte

func (s *Session) route(o types.Operation) handlerFunc {
	var handler handlerFunc
	switch o {
	case types.OperationHandshake:
		handler = s.handshake
	case types.OperationHeartbeat:
		handler = s.heartbeat
	case types.OperationConnect:
		handler = s.connect
	case types.OperationPush:
		handler = s.pushMessage
	case types.OperationNotification:
		handler = s.notification
	default:
		// Unknown operation
		log.Debug("[Dispatch] unknown operation", log.Ctx{"sid": s.sid})
		return nil
	}
	return handler
}

func (s *Session) dispatch(p *Protocol) {
	s.lastAction = time.Now().UTC()
	timestamp := s.lastAction.Unix()

	handler := s.route(p.Operation)
	if handler == nil {
		s.queueOut(p, ErrMalformed("", timestamp))
		return
	}

	message := &ServerMessage{
		Data:      p.Body,
		Timestamp: timestamp,
	}
	data := handler(message)
	if data != nil {
		s.queueOut(p, data)
		return
	}
}

func (s *Session) handshake(message *ServerMessage) []byte {
	var req HandshakeRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Handshake] failed to deserialize", log.Ctx{"error": err, "sid": s.sid})
		return ErrBadRequest("", message.Timestamp)
	}

	if s.version == 0 {
		version := x.ParseVersion(req.Version)
		if version == 0 {
			log.Debug("[Handshake] failed to parse version", "sid", s.sid)
			return ErrMalformed(req.MID, message.Timestamp)
		}
		// Check version compatibility
		if x.VersionCompare(version, minSupportedVersionValue) < 0 {
			log.Debug("[Handshake] unsupported version", "sid", s.sid)
			return ErrVersionNotSupported(req.MID, message.Timestamp)
		}

		clientID, id, err := s.srv.connect(s.ctx, req.Token, s.sid, s.serverID)
		if err != nil {
			log.Error("[Handshake] failed to connect", log.Ctx{"error": err, "sid": s.sid, "token": req.Token})
			return ErrInternalServer(req.MID, message.Timestamp, err.Error())
		}

		// Set user agent & platform in the beginning of the session.
		// Don't change them later.
		s.version = version
		s.userAgent = req.UserAgent
		s.platform = req.Platform
		if s.platform == "" {
			s.platform = x.PlatformFromUA(req.UserAgent)
		}

		// Only set uid in the first time authenticate.
		// Because uid is unique and will never change.
		if s.id.IsZero() {
			s.clientID = clientID
			s.id = id
		}
	}
	s.language = req.Language

	return NoErr(req.MID, message.Timestamp, nil)
}

func (s *Session) heartbeat(message *ServerMessage) []byte {
	var req HeartbeatRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Heartbeat] failed to deserialize", "sid", s.sid, "error", err)
		return ErrBadRequest("", message.Timestamp)
	}

	if s.id.IsZero() {
		return ErrAuthRequired(req.MID, message.Timestamp)
	}

	if err := s.srv.heartbeat(s.ctx, s.id.UID(), s.sid, s.serverID); err != nil {
		log.Error("[Heartbeat] failed to heartbeat", "sid", s.sid, "uid", s, s.id.UID(), "error", err)
		return ErrInternalServer(req.MID, message.Timestamp, err.Error())
	}

	return NoErr(req.MID, message.Timestamp, nil)
}

func (s *Session) connect(message *ServerMessage) []byte {
	var req ConnectRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Connect] failed to deserialize", "sid", s.sid, "error", err)
		return ErrBadRequest("", message.Timestamp)
	}

	clientID, uid, err := s.srv.connect(s.ctx, req.Token, s.sid, s.serverID)
	if err != nil {
		log.Error("[Connect] failed to connect", "sid", s.sid, "error", err)
		return ErrInternalServer(req.MID, message.Timestamp, err.Error())
	}

	// Only set uid in the first time authenticate.
	// Because uid is unique and will never change.
	if s.id.IsZero() {
		s.clientID = clientID
		s.id = uid
	}

	return NoErr(req.MID, message.Timestamp, nil)
}

func (s *Session) pushMessage(message *ServerMessage) []byte {
	var req PushMessageRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[PushMessage] failed to deserialize", "sid", s.sid, "error", err)
		return ErrBadRequest("", message.Timestamp)
	}

	if s.id.IsZero() {
		return ErrAuthRequired(req.MID, message.Timestamp)
	}

	messageID, sequence, err := s.srv.pushMessage(s.ctx, &chatApi.PushMessageReq{
		ClientID:    s.clientID,
		SID:         s.sid,
		MessageType: chatApi.MessageType(req.MessageType),
		Sender:      s.id.UID(),
		Receiver:    req.Receiver,
		ContentType: chatApi.ContentType(req.ContentType),
		Body:        req.Body,
		Mentions:    req.Mentions,
	})
	if err != nil {
		log.Error("[PushMessage] failed to push message", "sid", s.sid, "error", err)
		return ErrInternalServer(req.MID, message.Timestamp, err.Error())
	}

	resp := &PushMessageResponse{
		MessageID: messageID,
		Sequence:  sequence,
	}
	return NoErr(req.MID, message.Timestamp, resp)
}

func (s *Session) notification(message *ServerMessage) []byte {
	var req NotificationRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Notification] failed to deserialize", "sid", s.sid, "error", err)
		return ErrBadRequest("", message.Timestamp)
	}

	if s.id.IsZero() {
		return ErrAuthRequired(req.MID, message.Timestamp)
	}

	var err error
	switch req.What {
	case types.WhatTypeKeypress:
		err = s.srv.keypress(s.ctx, s.id.UID(), req.Topic)
	case types.WhatTypeRead:
		err = s.srv.readMessage(s.ctx, s.id.UID(), req.Topic, req.Sequence)
	}
	if err != nil {
		log.Error("[Notification] failed to send notification", "sid", s.sid, "error", err)
		return ErrInternalServer(req.MID, message.Timestamp, err.Error())
	}

	return NoErr(req.MID, message.Timestamp, nil)
}

func (s *Session) serialize(p *Protocol, body []byte) []byte {
	if p == nil {
		p = &Protocol{
			Operation: types.OperationUnknown,
		}
	}
	p.Body = body
	data, err := p.Marshal()
	if err != nil {
		log.Error("marshal error", "error", err)
		return nil
	}
	return data
}

func (s *Session) deserialize(v requester, body []byte) error {
	if v == nil {
		return ecode.NewError("requester can not be nil pointer")
	}
	if err := v.Unmarshal(body); err != nil {
		return err
	}
	if !v.Validate() {
		return ecode.NewError("failed to validate")
	}
	return nil
}

// queueOut attempts to send a ServerComMessage to a session; if the send buffer is full,
// timeout is `sendTimeout`.
func (s *Session) queueOut(p *Protocol, body []byte) bool {
	if s == nil {
		return true
	}
	select {
	case s.send <- s.serialize(p, body):
	case <-time.After(sendTimeout):
		log.Debug("[QueueOut] timeout", "sid", s.sid)
		return false
	}
	return true
}

func (s *Session) QueueOut(operation types.Operation, body []byte) bool {
	p := &Protocol{
		Operation: operation,
	}
	return s.queueOut(p, body)
}
