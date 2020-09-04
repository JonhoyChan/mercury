package persistence

import "outgoing/x/types"

type Message struct {
	ID          int64
	CreatedAt   int64
	Topic       string
	Sequence    int64
	MessageType types.MessageType
	Sender      int64
	Receiver    int64
	ContentType types.ContentType
	Body        []byte
	Status      uint8
	Mentions    []int64
}
