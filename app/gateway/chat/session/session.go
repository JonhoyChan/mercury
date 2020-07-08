package session

import (
	"context"
	"outgoing/app/gateway/chat/api"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"outgoing/x/types"
	"strings"
	"sync"
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
	// detach - channel for detaching session from topic, buffered.
	detach chan string

	// Map of topic subscriptions, indexed by topic name.
	// Don't access directly. Use getters/setters.
	subs map[string]*Subscription
	// Mutex for subs access: both topic go routines and network go routines access
	// subs concurrently.
	subsLock sync.RWMutex
}

// Subscription is a mapper of sessions to topics.
type Subscription struct {
	// Channel to communicate with the topic, copy of Topic.broadcast
	broadcast chan<- []byte

	//// Session sends a signal to Topic when this session is unsubscribed
	//// This is a copy of Topic.unreg
	//done chan<- *sessionLeave
	//
	//// Channel to send {meta} requests, copy of Topic.meta
	//meta chan<- *metaReq
	//
	//// Channel to ping topic with session's updates
	//supd chan<- *sessionUpdate
}

func (s *Session) addSub(topic string, sub *Subscription) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()
	s.subs[topic] = sub
}

func (s *Session) getSub(topic string) *Subscription {
	s.subsLock.RLock()
	defer s.subsLock.RUnlock()
	return s.subs[topic]
}

func (s *Session) delSub(topic string) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()
	delete(s.subs, topic)
}

func (s *Session) countSub() int {
	return len(s.subs)
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
			log.Error("ws: readLoop", log.Ctx{"error": err, "sid": s.sid})
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
				log.Warn("ws: outbound queue limit exceeded", log.Ctx{"sid": s.sid})
				return
			}
			if err := s.ws.WriteBinaryMessage(msg); err != nil {
				log.Error("ws: writeLoop", log.Ctx{"error": err, "sid": s.sid})
				return
			}
		case msg := <-s.stop:
			// Shutdown requested, don't care if the message is delivered
			if msg != nil {
				_ = s.ws.WriteTextMessage(msg)
			}
			return
		//case topic := <-s.detach:
		//	s.delSub(topic)
		case <-ticker.C:
			if err := s.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("ws: writeLoop ping", log.Ctx{"error": err, "sid": s.sid})
				return
			}
		}
	}
}

// Message received, convert bytes to ClientComMessage and dispatch
func (s *Session) dispatchRaw(raw []byte) {
	var (
		proto *api.Proto
		err   error
	)

	proto, err = api.Deserialize(raw)
	if err != nil {
		s.queueOut(s.serialize(proto, api.NewResponse(ecode.ErrBadRequest, "", "", 0)))
		return
	}

	s.dispatch(proto)
}

func (s *Session) dispatch(p *api.Proto) {
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
	case types.OperationSubscribe:
		handler = s.subscribe
	default:
		// Unknown operation
		log.Warn("[Dispatch] unknown operation", log.Ctx{"sid": s.sid})
		s.queueOut(s.serialize(p, api.NewResponse(ecode.ErrBadRequest, "", "", timestamp)))
		return
	}

	message := &ServerMessage{
		Data:      p.Body,
		Timestamp: timestamp,
	}
	resp := handler(message)
	if resp != nil {
		s.queueOut(s.serialize(p, resp))
	}
}

type marshaler interface {
	Marshal() ([]byte, error)
}

type handlerFunc func(message *ServerMessage) marshaler

func (s *Session) handshake(message *ServerMessage) marshaler {
	if message.Data == nil {
		log.Warn("[Handshake] proto body is nil", log.Ctx{"sid": s.sid})
		return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
	}

	var req api.HandshakeRequest
	if err := req.Unmarshal(message.Data); err != nil {
		log.Warn("[Handshake] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
		return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
	}

	if s.version == 0 {
		s.version = x.ParseVersion(req.Version)
		if s.version == 0 {
			log.Warn("[Handshake] failed to parse version", "sid", s.sid)
			return api.NewResponse(ecode.ErrBadRequest, req.MID, "", message.Timestamp)
		}
		// Check version compatibility
		if x.VersionCompare(s.version, minSupportedVersionValue) < 0 {
			s.version = 0
			log.Warn("[Handshake] unsupported version", "sid", s.sid)
			return api.NewResponse(ecode.ErrBadRequest.ResetMessage("unsupported version"), req.MID, "", message.Timestamp)
		}

		if req.Token == "" || req.UserAgent == "" {
			return api.NewResponse(ecode.ErrBadRequest, req.MID, "", message.Timestamp)
		}

		// Set user agent & platform in the beginning of the session.
		// Don't change them later.
		s.userAgent = req.UserAgent
		s.platform = req.Platform
		if s.platform == "" {
			s.platform = x.PlatformFromUA(req.UserAgent)
		}

		uid, authLevel, err := globalClient.authenticate(context.TODO(), req.Token, s.sid, s.serverID)
		if err != nil {
			log.Warn("[Authenticate] failed to authenticate token", log.Ctx{"error": err, "sid": s.sid, "token": ""})
			return api.NewResponse(ecode.ErrBadRequest, req.MID, "", message.Timestamp)
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
		//			Platform: s.platf,
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

	return api.NewResponse(ecode.Success, req.MID, "", message.Timestamp)
}

func (s *Session) heartbeat(message *ServerMessage) marshaler {
	var req api.HeartbeatRequest
	if message.Data != nil {
		if err := req.Unmarshal(message.Data); err != nil {
			log.Warn("[Heartbeat] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
			return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
		}
	}

	return api.NewResponse(ecode.Success, req.MID, "", message.Timestamp)
}

func (s *Session) authenticate(message *ServerMessage) marshaler {
	if message.Data == nil {
		log.Warn("[Authenticate] proto body is nil", log.Ctx{"sid": s.sid})
		return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
	}

	var req api.AuthenticateRequest
	if err := req.Unmarshal(message.Data); err != nil {
		log.Warn("[Authenticate] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
		return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
	}

	uid, authLevel, err := globalClient.authenticate(context.TODO(), req.Token, s.sid, s.serverID)
	if err != nil {
		log.Warn("[Authenticate] failed to authenticate token", log.Ctx{"error": err, "sid": s.sid, "token": ""})
		return api.NewResponse(ecode.ErrBadRequest, req.MID, "", message.Timestamp)
	}

	// Only set uid in the first time authenticate.
	// Because uid is unique and will never change.
	if s.uid.IsZero() {
		s.uid = uid
	}
	s.authLevel = authLevel

	return api.NewResponse(ecode.Success, req.MID, "", message.Timestamp)
}

// Request to subscribe to a topic
func (s *Session) subscribe(message *ServerMessage) marshaler {
	if message.Data == nil {
		log.Warn("[Subscribe] proto body is nil", log.Ctx{"sid": s.sid})
		return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
	}

	var req api.SubscribeRequest
	if err := req.Unmarshal(message.Data); err != nil {
		log.Warn("[Subscribe] failed to unmarshal", log.Ctx{"error": err, "sid": s.sid})
		return api.NewResponse(ecode.ErrBadRequest, "", "", message.Timestamp)
	}

	routeTo, resp := s.expandTopicName(req, message.Timestamp)
	if resp != nil {
		return resp
	}

	// Session can subscribe to topic on behalf of a single user at a time.
	if sub := s.getSub(routeTo); sub != nil {
		log.Warn("[Subscribe] already subscribed", log.Ctx{"sid": s.sid})
		return api.NewResponse(ecode.ErrDataAlreadyExist, req.MID, req.Topic, message.Timestamp)
	} else {
		globalHub.join <- &sessionJoin{
			mid:      req.MID,
			routeTo:  routeTo,
			original: req.Topic,
			session:  s,
		}
		// Hub will send success/failure packets back to session
	}

	return nil
}

// Expands session specific topic name to global name
func (s *Session) expandTopicName(req api.SubscribeRequest, timestamp int64) (string, marshaler) {
	if req.Topic == "" {
		log.Info("empty topic name", "sid", s.sid)
		return "", api.NewResponse(ecode.ErrBadRequest, req.MID, "", timestamp)
	}

	var routeTo string
	if req.Topic == "me" {
		routeTo = s.uid.UID()
	} else if req.Topic == "find" {
		routeTo = s.uid.PrefixId("find_")
	} else if strings.HasPrefix(req.Topic, "uid") {
		// p2p topic
		uid := types.ParseUid(req.Topic)
		if uid.IsZero() {
			// Ensure the user id is valid
			log.Info("failed to parse p2p topic name", "sid", s.sid)
			return "", api.NewResponse(ecode.ErrBadRequest, req.MID, req.Topic, timestamp)
		} else if s.uid == uid {
			// Use 'me' to access self-topic
			log.Info("invalid p2p self-subscription", "sid", s.sid)
			return "", api.NewResponse(ecode.ErrUnauthorized, req.MID, req.Topic, timestamp)
		}

		routeTo = s.uid.P2PName(uid)
	} else {
		routeTo = req.Topic
	}

	return routeTo, nil
}

func (s *Session) serialize(p *api.Proto, v marshaler) []byte {
	if p == nil {
		p = api.NewProto(int32(types.OperationUnknown))
	}
	p.Body, _ = v.Marshal()
	data, _ := api.Serialize(p)
	return data
}

// queueOut attempts to send a ServerComMessage to a session; if the send buffer is full,
// timeout is `sendTimeout`.
func (s *Session) queueOut(data []byte) bool {
	if s == nil {
		return true
	}
	select {
	case s.send <- data:
	case <-time.After(sendTimeout):
		log.Debug("ws.queueOut: timeout", log.Ctx{"sid": s.sid})
		return false
	}
	return true
}
