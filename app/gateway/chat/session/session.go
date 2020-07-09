package session

import (
	"context"
	"encoding/json"
	"outgoing/app/gateway/chat/api"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"outgoing/x/types"
	"time"

	"outgoing/x/websocket"
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
	// currentVersion is the current version
	CurrentVersion = "0.1"
	// minSupportedVersion is the minimum supported version
	MinSupportedVersion = "0.1"
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
	// ID of the current user or 0.
	uid types.Uid
	// Authentication level - NONE (unset), ANON, AUTH, ROOT.
	authLevel types.AuthLevel
	// Time when the session received any packer from client.
	lastAction time.Time
	// Outbound messages, buffered.
	// The content must be serialized in format suitable for the session.
	send chan []byte
	// Channel for shutting down the session, buffer 1.
	// Content in the same format as for 'send'.
	stop chan []byte
}

func (s *Session) readLoop() {
	defer func() {
		s.ws.Close()
		GlobalSessionStore.Delete(s)
	}()

	s.ws.SetReadLimit(MaxMessageSize)
	s.ws.SetReadDeadline(time.Now().Add(pongWait))
	s.ws.SetPongHandler(func(string) error {
		s.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read a ClientComMessage
		raw, err := s.ws.ReadMessage()
		if err != nil {
			log.Error("[Websocket] failed to read message", log.Ctx{"error": err, "sid": s.sid})
			return
		}
		// TODO Remove this log output
		log.Debug(string(raw))

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

type Proto struct {
	// protocol version
	Version int32 `json:"version"`
	// operation for request
	Operation int32 `json:"operation"`
	// binary body bytes
	Body json.RawMessage `json:"body"`
}

// Message received, convert bytes to ClientComMessage and dispatch
func (s *Session) dispatchRaw(raw []byte) {
	var (
		//proto *api.Proto
		err error
	)

	var jsonProto Proto
	if err = json.Unmarshal(raw, &jsonProto); err != nil {
		s.queueOut(&jsonProto, ErrMalformed("", 0))
		return
	}
	//proto, err = api.Deserialize(raw)
	//if err != nil {
	//	s.queueOut(proto, ErrMalformed("", 0))
	//	return
	//}

	s.dispatch(&jsonProto)
}

type unmarshaler interface {
	Unmarshal([]byte) error
}

type handlerFunc func(message *ServerMessage) []byte

func (s *Session) dispatch(p *Proto) {
	s.lastAction = time.Now().UTC()
	timestamp := s.lastAction.Unix()

	var handler handlerFunc
	switch types.Operation(p.Operation) {
	case types.OperationHandshake:
		handler = s.handshake
	case types.OperationHeartbeat:
		handler = s.heartbeat
	case types.OperationAuthenticate:
		handler = s.authenticate
	default:
		// Unknown operation
		log.Debug("[Dispatch] unknown operation", log.Ctx{"sid": s.sid})
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
	if message.Data == nil {
		log.Debug("[Handshake] proto body is nil", log.Ctx{"sid": s.sid})
		return ErrMalformed("", message.Timestamp)
	}

	var req api.HandshakeRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Handshake] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
		return ErrMalformed("", message.Timestamp)
	}

	if s.version == 0 {
		s.version = x.ParseVersion(req.Version)
		if s.version == 0 {
			log.Debug("[Handshake] failed to parse version", "sid", s.sid)
			return ErrMalformed(req.MID, message.Timestamp)
		}
		// Check version compatibility
		if x.VersionCompare(s.version, minSupportedVersionValue) < 0 {
			s.version = 0
			log.Debug("[Handshake] unsupported version", "sid", s.sid)
			return ErrVersionNotSupported(req.MID, message.Timestamp)
		}

		if req.Token == "" || req.UserAgent == "" {
			return ErrMalformed(req.MID, message.Timestamp)
		}

		// Set user agent & platform in the beginning of the session.
		// Don't change them later.
		s.userAgent = req.UserAgent
		s.platform = req.Platform
		if s.platform == "" {
			s.platform = x.PlatformFromUA(req.UserAgent)
		}

		uid, authLevel, err := globalClient.authenticate(s.ctx, req.Token, s.sid, s.serverID)
		if err != nil {
			log.Error("[Authenticate] failed to authenticate token", log.Ctx{"error": err, "sid": s.sid, "token": ""})
			return ErrAuthFailed(req.MID, message.Timestamp)
		}

		// Only set uid in the first time authenticate.
		// Because uid is unique and will never change.
		if s.uid.IsZero() {
			s.uid = uid
		}
		s.authLevel = authLevel
	} else if x.ParseVersion(req.Version) == s.version {
		//// Save changed device ID+Lang or delete earlier specified device ID.
		//// Platform cannot be changed.
		//if !s.uid.IsZero() {
		//	var err error
		//	if msg.Hi.DeviceID == types.NullValue {
		//		deviceIDUpdate = true
		//		err = store.Devices.Delete(s.uid, s.deviceID)
		//	} else if msg.Hi.DeviceID != "" {
		//		deviceIDUpdate = true
		//		err = store.Devices.Update(s.uid, s.deviceID, &types.DeviceDef{
		//			DeviceId: msg.Hi.DeviceID,
		//			Platform: s.platform,
		//			LastSeen: msg.timestamp,
		//			Lang:     msg.Hi.Lang,
		//		})
		//	}
		//
		//	if err != nil {
		//		log.Println("s.hello:", "device ID", err, s.sid)
		//		s.queueOut(ErrUnknown(msg.Id, "", msg.timestamp))
		//		return
		//	}
		//}
	}

	if req.DeviceID == types.NullValue {
		req.DeviceID = ""
	}
	s.deviceID = req.DeviceID
	s.language = req.Language

	return NoErr(req.MID, message.Timestamp)
}

func (s *Session) heartbeat(message *ServerMessage) []byte {
	if message.Data == nil {
		log.Debug("[Heartbeat] proto body is nil", log.Ctx{"sid": s.sid})
		return ErrMalformed("", message.Timestamp)
	}

	var req api.HeartbeatRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Heartbeat] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
		return ErrMalformed("", message.Timestamp)
	}

	if s.uid.IsZero() {
		log.Debug("[Heartbeat] uid is zero", log.Ctx{"sid": s.sid})
		return ErrAuthRequired(req.MID, message.Timestamp)
	}

	if err := globalClient.heartbeat(s.ctx, s.uid.UID(), s.sid, s.serverID); err != nil {
		log.Error("[Heartbeat] failed to heartbeat", log.Ctx{"error": err, "uid": s.uid.UID(), "sid": s.sid})
		return ErrInternalServer(req.MID, message.Timestamp)
	}

	return NoErr(req.MID, message.Timestamp)
}

func (s *Session) authenticate(message *ServerMessage) []byte {
	if message.Data == nil {
		log.Debug("[Authenticate] proto body is nil", log.Ctx{"sid": s.sid})
		return ErrMalformed("", message.Timestamp)
	}

	var req api.AuthenticateRequest
	if err := s.deserialize(&req, message.Data); err != nil {
		log.Warn("[Authenticate] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
		return ErrMalformed("", message.Timestamp)
	}

	uid, authLevel, err := globalClient.authenticate(s.ctx, req.Token, s.sid, s.serverID)
	if err != nil {
		log.Error("[Authenticate] failed to authenticate token", log.Ctx{"error": err, "sid": s.sid, "token": ""})
		return ErrAuthFailed(req.MID, message.Timestamp)
	}

	// Only set uid in the first time authenticate.
	// Because uid is unique and will never change.
	if s.uid.IsZero() {
		s.uid = uid
	}
	s.authLevel = authLevel

	return NoErr(req.MID, message.Timestamp)
}

func (s *Session) serialize(p *Proto, body []byte) []byte {
	if p == nil {
		//p = api.NewProto(int32(types.OperationUnknown))
		p = &Proto{
			Version:   api.ProtocolVersion,
			Operation: int32(types.OperationUnknown),
		}
	}
	p.Body = body
	//data, _ := api.Serialize(p)
	data, _ := json.Marshal(&p)
	return data
}

func (s *Session) deserialize(v unmarshaler, body []byte) error {
	if v == nil {
		return ecode.NewError("unmarshaler can not be nil")
	}
	//return v.Unmarshal(body)
	return json.Unmarshal(body, v)
}

// queueOut attempts to send a ServerComMessage to a session; if the send buffer is full,
// timeout is `sendTimeout`.
func (s *Session) queueOut(p *Proto, body []byte) bool {
	if s == nil {
		return true
	}
	select {
	case s.send <- s.serialize(p, body):
	case <-time.After(sendTimeout):
		log.Debug("[QueueOut] timeout", log.Ctx{"sid": s.sid})
		return false
	}
	return true
}
