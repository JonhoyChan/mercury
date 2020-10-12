package service

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client"
	cApi "mercury/app/comet/api"
	"mercury/x/log"
	"sync/atomic"
)

type Comet struct {
	ctx              context.Context
	cancel           context.CancelFunc
	serverID         string
	callOption       client.CallOption
	pushChan         []chan *cApi.PushMessageReq
	broadcastChan    []chan *cApi.BroadcastMessageReq
	pushChanNum      uint64
	broadcastChanNum uint64
	routineSize      uint64
}

func NewComet(id, address string) (*Comet, error) {
	if address == "" {
		return nil, fmt.Errorf("invalid node address: %v", address)
	}

	routineSize := 32
	c := &Comet{
		serverID:      id,
		callOption:    client.WithAddress(address),
		pushChan:      make([]chan *cApi.PushMessageReq, routineSize),
		broadcastChan: make([]chan *cApi.BroadcastMessageReq, routineSize),
		routineSize:   uint64(routineSize),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	for i := 0; i < routineSize; i++ {
		c.pushChan[i] = make(chan *cApi.PushMessageReq, 1024)
		go c.process(c.pushChan[i], c.broadcastChan[i])
	}
	return c, nil
}

func (c *Comet) Push(req *cApi.PushMessageReq) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- req
}

func (c *Comet) Broadcast(req *cApi.BroadcastMessageReq) {
	idx := atomic.AddUint64(&c.broadcastChanNum, 1) % c.routineSize
	c.broadcastChan[idx] <- req
}

func (c *Comet) process(pushChan chan *cApi.PushMessageReq, broadcastChan chan *cApi.BroadcastMessageReq) {
	for {
		select {
		case req := <-pushChan:
			_, err := grpcClient.PushMessage(c.ctx, req, c.callOption)
			if err != nil {
				log.Error("failed to push message", "error", err)
			}
		case req := <-broadcastChan:
			_, err := grpcClient.BroadcastMessage(c.ctx, req, c.callOption)
			if err != nil {
				log.Error("failed to broadcast message", "error", err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}
