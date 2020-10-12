package service

import (
	"mercury/app/logic/api"
	"mercury/x/types"
)

func (s *Service) MessageListenerCount(clientID string) int {
	return len(s.messageChan[clientID])
}

func (s *Service) InvokeMessageListener(clientID string, message *types.Message) {
	c, ok := s.messageChan[clientID]
	if ok {
		c <- message
	}
}

func (s *Service) listen(clientID string, stream api.ChatAdmin_ListenStream) error {
	_, ok := s.messageChan[clientID]
	if !ok {
		s.messageChan[clientID] = make(chan *types.Message, 4096)
	}
	for {
		select {
		case <-s.doneChan:
			return stream.Close()
		case m := <-s.messageChan[clientID]:
			if err := stream.Send(&api.Message{
				ID:          m.ID,
				CreatedAt:   m.CreatedAt,
				MessageType: m.MessageType.String(),
				Sender:      m.Sender,
				Receiver:    m.Receiver,
				Topic:       m.Topic,
				Sequence:    m.Sequence,
				ContentType: m.ContentType.String(),
				Body:        m.Body,
				Mentions:    m.Mentions,
			}); err != nil {
				return err
			}
		}
	}
}
