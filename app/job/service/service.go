package service

import (
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	cApi "mercury/app/comet/api"
	"mercury/config"
	"mercury/x/ecode"
	"mercury/x/log"
	"strings"
	"sync"
	"time"
)

type Servicer interface {
	Init(options server.Options)
	Close()
}

var grpcClient cApi.ChatService

type Service struct {
	broker       broker.Broker
	config       ConfigProvider
	cometServers map[string]*Comet
	log          log.Logger
	mutex        *sync.Mutex
	registry     registry.Registry
	stopChan     chan struct{}
	watchChan    chan bool
}

type ConfigProvider interface {
	Topic() config.Topic
}

func NewService(config ConfigProvider, l log.Logger) (*Service, error) {
	return &Service{
		config:       config,
		cometServers: make(map[string]*Comet),
		log:          l,
		mutex:        &sync.Mutex{},
		stopChan:     make(chan struct{}),
		watchChan:    make(chan bool, 1),
	}, nil
}

func (s *Service) Init(options server.Options) {
	s.withRegistry(options.Registry)
	s.withBroker(options.Broker)
}

func (s *Service) withRegistry(r registry.Registry) {
	if s.registry == nil {
		s.registry = r

		opts := []client.Option{
			client.RequestTimeout(10 * time.Second),
			client.Retries(2),
			client.Retry(ecode.RetryOnMicroError),
			client.WrapCall(ecode.MicroCallFunc),
			client.Registry(r),
		}

		c := grpc.NewClient(opts...)

		grpcClient = cApi.NewChatService("mercury.comet", c)

		go s.watchComet()
	}
}

func (s *Service) withBroker(b broker.Broker) {
	if s.broker == nil {
		s.broker = b

		topic := s.config.Topic()
		pushMessageTopic, ok := topic.Get("push_message")
		if ok {
			if _, err := s.broker.Subscribe(pushMessageTopic, s.subscribePushMessage); err != nil {
				s.log.Error("[WatchComet] failed to subscribe topic", "topic", pushMessageTopic, "error", err)
				return
			}
		}
		broadcastMessageTopic, ok := topic.Get("broadcast_message")
		if ok {
			if _, err := s.broker.Subscribe(broadcastMessageTopic, s.subscribeBroadcastMessage); err != nil {
				s.log.Error("[WatchComet] failed to subscribe topic", "topic", broadcastMessageTopic, "error", err)
				return
			}
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

func (s *Service) watchComet() {
	if err := s.syncCometNodes(); err != nil {
		panic("failed to sync comet nodes:" + err.Error())
	}

	go s.watch()
	go s.sync()
}

func (s *Service) watch() {
	cometServiceName := "mercury.comet"
	watcher, err := s.registry.Watch(registry.WatchService(cometServiceName))
	if err != nil {
		panic("failed to watch service:" + err.Error())
	}

	for {
		result, err := watcher.Next()
		if err != nil {
			if err != registry.ErrWatcherStopped {
				s.log.Error("[watch] failed to watch next", "error", err)
			} else {
				s.log.Error("[watch] watcher stopped")
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
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.syncCometNodes(); err != nil {
				s.log.Error("[sync] failed to sync comet nodes", "error", err)
			}
		case <-s.watchChan:
			if err := s.syncCometNodes(); err != nil {
				s.log.Error("[sync] failed to sync comet nodes", "error", err)
			}
		case <-s.stopChan:
			s.log.Info("[sync] sync stopped")
			return
		}
	}
}

func (s *Service) syncCometNodes() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	cometServiceName := "mercury.comet"
	cometServices, err := s.registry.GetService(cometServiceName)
	if err != nil {
		s.log.Error("[syncCometNodes] failed to new comet", "error", err)
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
				s.log.Error("[syncCometNodes] can not new comet", "error", err)
				return err
			}

			comets[id] = c

			s.log.Info("[syncCometNodes] new comet", "id", id, "address", node.Address)
		}

		for id, old := range s.cometServers {
			if _, ok := comets[id]; !ok {
				old.cancel()
				s.log.Info("[syncCometNodes] delete comet", "id", id)
			}
		}

		s.cometServers = comets
	}
	return nil
}
