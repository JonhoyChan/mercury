package service

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client"
)

type Comet struct {
	ctx        context.Context
	cancel     context.CancelFunc
	serverID   string
	callOption client.CallOption
	//pushChan      []chan *api.PushMsgReq
	//roomChan      []chan *api.BroadcastRoomReq
	//broadcastChan chan *api.BroadcastReq
	//pushChanNum   uint64
	//roomChanNum   uint64
	//routineSize   uint64
}

func NewComet(id, address string) (*Comet, error) {
	if address == "" {
		return nil, fmt.Errorf("invalid node address: %v", address)
	}

	comet := &Comet{
		serverID:   id,
		callOption: client.WithAddress(address),
		//pushChan:      make([]chan *comet.PushMsgReq, c.RoutineSize),
		//roomChan:      make([]chan *comet.BroadcastRoomReq, c.RoutineSize),
		//broadcastChan: make(chan *comet.BroadcastReq, c.RoutineSize),
		//routineSize:   uint64(c.RoutineSize),
	}
	comet.ctx, comet.cancel = context.WithCancel(context.Background())

	//for i := 0; i < c.RoutineSize; i++ {
	//	cmt.pushChan[i] = make(chan *comet.PushMsgReq, c.RoutineChan)
	//	cmt.roomChan[i] = make(chan *comet.BroadcastRoomReq, c.RoutineChan)
	//	go cmt.process(cmt.pushChan[i], cmt.roomChan[i], cmt.broadcastChan)
	//}
	return comet, nil
}
