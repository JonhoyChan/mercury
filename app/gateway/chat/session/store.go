/******************************************************************************
 *
 *  Description:
 *
 *  Session management.
 *
 *****************************************************************************/

package session

import (
	"context"
	"outgoing/app/gateway/chat/stats"
	"outgoing/x"
	"outgoing/x/ksuid"
	"outgoing/x/log"
	"outgoing/x/websocket"
)

var GlobalSessionStore = NewSessionStore()

// SessionStore holds live sessions. Long polling sessions are stored in a linked list with
// most recent sessions on top. In addition all sessions are stored in a map indexed by session ID.
type SessionStore struct {
	cache Cache
}

// NewSessionStore initializes a session store.
func NewSessionStore() *SessionStore {
	ss := &SessionStore{
		cache: NewDefaultCache(),
	}

	stats.RegisterInt("LiveSessions")
	stats.RegisterInt("TotalSessions")

	return ss
}

// NewSession creates a new session and saves it to the session store.
func (ss *SessionStore) NewSession(ctx context.Context, conn interface{}, serverID string) {
	var s Session
	s.ctx = ctx
	s.sid = ksuid.New().String()
	s.serverID = serverID

	if ss.cache.Existed(s.sid) {
		panic(x.Sprintf("duplicate session ID", s.sid))
	}

	switch c := conn.(type) {
	case *websocket.Connection:
		s.proto = WEBSOCKET
		s.ws = c
	default:
		s.proto = NONE
	}

	if s.proto != NONE {
		//s.subs = make(map[string]*Subscription)
		s.send = make(chan []byte, sendQueueLimit+32) // buffered
		s.stop = make(chan []byte, 1)                 // Buffered by 1 just to make it non-blocking
	}

	ss.cache.Store(s.sid, &s)

	if s.proto == WEBSOCKET {
		// Do work in goroutines to return from serveWebSocket() to release file pointers.
		// Otherwise "too many open files" will happen.
		go s.readLoop()
		go s.writeLoop()

		log.Info("[Websocket] session stored", "sid", s.sid, "count", ss.cache.Length())
	}
}

// Get fetches a session from store by session ID.
func (ss *SessionStore) Get(sid string) *Session {
	return ss.cache.Load(sid)
}

// Delete removes session from store.
func (ss *SessionStore) Delete(s *Session) {
	ss.cache.Delete(s.sid)
	if s.proto == WEBSOCKET {
		log.Info("[Websocket] session deleted", "sid", s.sid, "count", ss.cache.Length())
	}
}

// Shutdown terminates sessionStore. No need to clean up.
// Don't send to clustered sessions, their servers are not being shut down.
func (ss *SessionStore) Shutdown() {
	ss.cache.Shutdown()
	log.Debug(x.Sprintf("[SessionStore] shut down. %d sessions terminated", ss.cache.Length()))
}
