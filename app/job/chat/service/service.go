package service

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/registry"
	"outgoing/app/job/chat/config"
	cApi "outgoing/app/service/chat/api"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"time"
)

type Service struct {
	config       config.Provider
	registry     registry.Registry
	cometServers map[string]*Comet
	watchChan    chan bool
	stopChan     chan struct{}
	// TODO change to chat gateway client
	chatService cApi.ChatService
}

func NewService(config config.Provider) *Service {
	opts := []client.Option{
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Service{
		config:      config,
		watchChan:   make(chan bool, 1),
		stopChan:    make(chan struct{}),
		chatService: cApi.NewChatService("gate.srv", c),
	}
}

func (s *Service) WithRegistry(registry registry.Registry) {
	if s.registry == nil {
		s.registry = registry
	}
}

func (s *Service) Close() {
	for id, old := range s.cometServers {
		old.cancel()
		log.Info("[Close] job server close", "id", id)
	}
	s.stopChan <- struct{}{}
}

func (s *Service) WatchComet() {
	if err := s.syncCometNodes(); err != nil {
		panic("failed to sync comet nodes:" + err.Error())
	}

	go s.watch()
	go s.sync()
}

func (s *Service) watch() {
	watcher, err := s.registry.Watch(registry.WatchService("gateway.chat.comet"))
	if err != nil {
		panic("failed to watch service:" + err.Error())
	}

	for {
		result, err := watcher.Next()
		if err != nil {
			if err != registry.ErrWatcherStopped {
				log.Error("[WatchComet] failed to watch next", "error", err)
			} else {
				log.Error("[WatchComet] watcher stopped")
			}
			break
		}

		if result.Action != "create" {
			continue
		}

		s.watchChan <- true
	}
}

func (s *Service) sync() {
	ticker := time.NewTicker(s.config.RegisterInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.syncCometNodes(); err != nil {
				log.Error("[Sync] failed to sync comet nodes", "error", err)
			}
		case <-s.watchChan:
			if err := s.syncCometNodes(); err != nil {
				log.Error("[Sync] failed to sync comet nodes", "error", err)
			}
		case <-s.stopChan:
			log.Info("[Sync] sync stopped")
			return
		}
	}
}

func (s *Service) syncCometNodes() error {
	cometServices, err := s.registry.GetService("gateway.chat.comet")
	if err != nil {
		log.Error("[SyncCometNodes] failed to new comet", "error", err)
		return err
	}

	var nodes []*registry.Node
	for _, cometService := range cometServices {
		nodes = append(nodes, cometService.Nodes...)
	}

	comets := make(map[string]*Comet)
	for _, node := range nodes {
		if old, ok := s.cometServers[node.Id]; ok {
			comets[node.Id] = old
			continue
		}

		c, err := NewComet(node)
		if err != nil {
			log.Error("[SyncCometNodes] can not new comet", "error", err)
			return err
		}

		comets[node.Id] = c

		log.Info("[SyncCometNodes] new comet", "id", node.Id, "address", node.Address)
	}

	for id, old := range s.cometServers {
		if _, ok := comets[id]; !ok {
			old.cancel()
			log.Info("[SyncCometNodes] delete comet", "id", id)
		}
	}

	s.cometServers = comets
	return nil
}

//func (s *Service) PushMessage(serverID string) error {
//	comet, ok := s.cometServers[serverID]
//	if !ok {
//		return
//	}
//
//	s.chatService.Connect(comet.ctx, &cApi.ConnectReq{}, comet.callOption)
//}
