package session

import (
	"outgoing/app/gateway/chat/api"
	"outgoing/app/gateway/chat/stats"
	"outgoing/x/log"
	"outgoing/x/types"
	"sync"
	"time"
)

const (
	// idleSessionTimeout defines duration of being idle before terminating a session.
	idleSessionTimeout = time.Second * 55
)

// Request to hub to subscribe session to topic
type sessionJoin struct {
	mid      string
	routeTo  string
	original string
	proto    *api.Proto
	session  *Session
}

// Session wants to leave the topic
type sessionLeave struct {
	// Message, containing request details. Could be nil.
	data []byte
	// Session which initiated the request
	session *Session
}

var globalHub = newHub()

// Hub is the core structure which holds topics.
type Hub struct {
	// Topics must be indexed by name
	topics *sync.Map
	// subscribe session to topic, possibly creating a new topic, unbuffered
	join chan *sessionJoin
	// Request to shutdown, unbuffered
	shutdown chan chan<- bool
}

func (h *Hub) topicGet(name string) *Topic {
	if t, ok := h.topics.Load(name); ok {
		return t.(*Topic)
	}
	return nil
}

func (h *Hub) topicPut(name string, t *Topic) {
	h.topics.Store(name, t)
}

func (h *Hub) topicDel(name string) {
	h.topics.Delete(name)
}

func newHub() *Hub {
	var h = &Hub{
		topics:   &sync.Map{},
		join:     make(chan *sessionJoin),
		shutdown: make(chan chan<- bool),
	}

	stats.RegisterInt("LiveTopics")
	stats.RegisterInt("TotalTopics")

	go h.run()

	return h
}

func (h *Hub) run() {
	for {
		select {
		case join := <-h.join:
			// Is the topic already loaded?
			t := h.topicGet(join.routeTo)
			if t == nil {
				// Topic does not exist or not loaded.
				t = &Topic{
					name:      join.routeTo,
					original:  join.original,
					sessions:  make(map[*Session]perSessionData),
					broadcast: make(chan []byte, 256),
					join:      make(chan *sessionJoin, 32),
					leave:     make(chan *sessionLeave, 32),
					//meta:      make(chan *metaReq, 32),
					perUser: make(map[types.Uid]perUserData),
					exit:    make(chan *shutDown, 1),
				}
				// Topic is created in suspended state because it's not yet configured.
				t.markPaused(true)
				// Save topic now to prevent race condition.
				h.topicPut(join.routeTo, t)

				// Configure the topic.
				go topicInit(t, join, h)

			} else {
				// Topic found.
				// Topic will check access rights and send appropriate message
				t.join <- join
			}
		case done := <-h.shutdown:
			// start cleanup process
			topicsDone := make(chan bool)
			topicCount := 0
			h.topics.Range(func(_, v interface{}) bool {
				v.(*Topic).exit <- &shutDown{done: topicsDone}
				topicCount++
				return true
			})

			for i := 0; i < topicCount; i++ {
				<-topicsDone
			}

			log.Info("[Hub] shutdown completed", log.Ctx{"topics": topicCount})

			// let the main goroutine know we are done with the cleanup
			done <- true

			return

		case <-time.After(idleSessionTimeout):
		}
	}
}
