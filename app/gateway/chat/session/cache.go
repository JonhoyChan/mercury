package session

import (
	"outgoing/app/gateway/chat/api"
	"outgoing/app/gateway/chat/stats"
	"outgoing/x/ecode"
	"outgoing/x/types"
	"sync"
)

type Cache interface {
	Length() int
	Existed(key string) bool
	Store(key string, value *Session)
	Load(key string) *Session
	Delete(key string)
	Shutdown()
	EvictUser(uid types.Uid, skipSid string)
}

func NewDefaultCache() Cache {
	return &defaultCache{
		mux: sync.RWMutex{},
		kv:  make(map[string]*Session),
	}
}

type defaultCache struct {
	mux sync.RWMutex
	// All sessions indexed by session ID
	kv map[string]*Session
}

func (c *defaultCache) Length() int {
	return len(c.kv)
}

func (c *defaultCache) Existed(key string) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	_, ok := c.kv[key]
	return ok
}

func (c *defaultCache) Store(key string, value *Session) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.kv[key] = value

	stats.Set("LiveSessions", len(c.kv), false)
	stats.Set("TotalSessions", 1, true)
}

func (c *defaultCache) Load(key string) *Session {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.kv[key]
}

func (c *defaultCache) Delete(key string) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	delete(c.kv, key)

	stats.Set("LiveSessions", len(c.kv), false)
}

func (c *defaultCache) Shutdown() {
	for _, s := range c.kv {
		s.stop <- s.serialize(nil, api.NewResponse(ecode.ResetContent.ResetMessage("server shutdown"),
			"", "", 0))
	}
}

func (c *defaultCache) EvictUser(uid types.Uid, skipSid string) {
	for _, s := range c.kv {
		if s.uid == uid && s.stop != nil && s.sid != skipSid {
			s.stop <- s.serialize(nil, api.NewResponse(ecode.ResetContent.ResetMessage("evicted"),
				"", "", 0))
			delete(c.kv, s.sid)
		}
	}

	stats.Set("LiveSessions", len(c.kv), false)
}
