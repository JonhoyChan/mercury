package service

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client"
	cApi "outgoing/app/gateway/chat/api"
	"outgoing/x/log"
	"sync/atomic"
)

type Comet struct {
	ctx         context.Context
	cancel      context.CancelFunc
	serverID    string
	callOption  client.CallOption
	pushChan    []chan *cApi.PushMessageReq
	pushChanNum uint64
	routineSize uint64
}

func NewComet(id, address string) (*Comet, error) {
	if address == "" {
		return nil, fmt.Errorf("invalid node address: %v", address)
	}

	routineSize := 32
	c := &Comet{
		serverID:    id,
		callOption:  client.WithAddress(address),
		pushChan:    make([]chan *cApi.PushMessageReq, routineSize),
		routineSize: uint64(routineSize),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	for i := 0; i < routineSize; i++ {
		c.pushChan[i] = make(chan *cApi.PushMessageReq, 1024)
		go c.process(c.pushChan[i])
	}
	return c, nil
}

func (c *Comet) Push(req *cApi.PushMessageReq) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- req
}

func (c *Comet) process(pushChan chan *cApi.PushMessageReq) {
	for {
		select {
		case req := <-pushChan:
			_, err := grpcClient.PushMessage(c.ctx, &cApi.PushMessageReq{SIDs: req.SIDs, Data: req.Data}, c.callOption)
			if err != nil {
				log.Error("failed to push message", "error", err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}
