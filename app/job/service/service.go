package service

import (
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/registry"
	cApi "mercury/app/comet/api"
	"mercury/app/job/config"
	"mercury/x/ecode"
	"mercury/x/log"
	"strings"
	"sync"
	"time"
)

var grpcClient cApi.ChatService

type Service struct {
	mutex        *sync.Mutex
	config       config.Provider
	registry     registry.Registry
	broker       broker.Broker
	cometServers map[string]*Comet
	watchChan    chan bool
	stopChan     chan struct{}
}

func NewService(config config.Provider) *Service {
	return &Service{
		mutex:        &sync.Mutex{},
		config:       config,
		cometServers: make(map[string]*Comet),
		watchChan:    make(chan bool, 1),
		stopChan:     make(chan struct{}),
	}
}

func (s *Service) WithRegistry(r registry.Registry) {
	if s.registry == nil {
		s.registry = r

		opts := []client.Option{
			client.Retries(2),
			client.Retry(ecode.RetryOnMicroError),
			client.WrapCall(ecode.MicroCallFunc),
			client.Registry(r),
		}

		c := grpc.NewClient(opts...)

		grpcClient = cApi.NewChatService(s.config.CometServiceName(), c)
	}
}

func (s *Service) WithBroker(b broker.Broker) {
	if s.broker == nil {
		s.broker = b

		if _, err := s.broker.Subscribe(s.config.PushMessageTopic(), s.subscribePushMessage); err != nil {
			log.Error("[WatchComet] failed to subscribe topic", "topic", s.config.PushMessageTopic(), "error", err)
			return
		}
		if _, err := s.broker.Subscribe(s.config.BroadcastMessageTopic(), s.subscribeBroadcastMessage); err != nil {
			log.Error("[WatchComet] failed to subscribe topic", "topic", s.config.BroadcastMessageTopic(), "error", err)
			return
		}
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
	if s.registry == nil {
		panic("registry is nil, please use the WithRegistry first")
	}

	if err := s.syncCometNodes(); err != nil {
		panic("failed to sync comet nodes:" + err.Error())
	}

	go s.watch()
	go s.sync()
}

func (s *Service) watch() {
	cometServiceName := s.config.CometServiceName()
	watcher, err := s.registry.Watch(registry.WatchService(cometServiceName))
	if err != nil {
		panic("failed to watch service:" + err.Error())
	}

	for {
		result, err := watcher.Next()
		if err != nil {
			if err != registry.ErrWatcherStopped {
				log.Error("[watch] failed to watch next", "error", err)
			} else {
				log.Error("[watch] watcher stopped")
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
				log.Error("[sync] failed to sync comet nodes", "error", err)
			}
		case <-s.watchChan:
			if err := s.syncCometNodes(); err != nil {
				log.Error("[sync] failed to sync comet nodes", "error", err)
			}
		case <-s.stopChan:
			log.Info("[sync] sync stopped")
			return
		}
	}
}

func (s *Service) syncCometNodes() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	cometServiceName := s.config.CometServiceName()
	cometServices, err := s.registry.GetService(cometServiceName)
	if err != nil {
		log.Error("[syncCometNodes] failed to new comet", "error", err)
		return err
	}

	if len(cometServices) > 0 {
		var nodes []*registry.Node
		for _, cometService := range cometServices {
			nodes = append(nodes, cometService.Nodes...)
		}

		comets := make(map[string]*Comet)
		for _, node := range nodes {
			if !strings.HasPrefix(node.Id, cometServiceName) {
				continue
			}

			id := strings.TrimPrefix(node.Id, cometServiceName+"-")
			if old, ok := s.cometServers[id]; ok {
				comets[id] = old
				continue
			}

			c, err := NewComet(id, node.Address)
			if err != nil {
				log.Error("[syncCometNodes] can not new comet", "error", err)
				return err
			}

			comets[id] = c

			log.Info("[syncCometNodes] new comet", "id", id, "address", node.Address)
		}

		for id, old := range s.cometServers {
			if _, ok := comets[id]; !ok {
				old.cancel()
				log.Info("[syncCometNodes] delete comet", "id", id)
			}
		}

		s.cometServers = comets
	}
	return nil
}
