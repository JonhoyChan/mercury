package ipfs

import (
	//"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	u "github.com/ipfs/go-ipfs-util"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p-pubsub"
)

const messageTopicPrefix = "/offlinemessage/"

type Pubsub struct {
	Subscriber *PubsubSubscriber
	Publisher  *PubsubPublisher
}

// PubsubPublisher is a publisher that distributes arbitrary data through pubsub
type PubsubPublisher struct {
	ctx  context.Context
	ds   datastore.Datastore
	host host.Host
	cr   routing.ContentRouting
	ps   *pubsub.PubSub

	mx   sync.Mutex
	subs map[string]struct{}
}

// PubsubSubscriber subscribes to arbitrary subscriptions through pubsub
type PubsubSubscriber struct {
	ctx  context.Context
	ds   datastore.Datastore
	host host.Host
	cr   routing.ContentRouting
	ps   *pubsub.PubSub

	mx   sync.Mutex
	subs map[string]*pubsub.Subscription
}

// NewPubsubPublisher constructs a new Publisher that publishes arbitrary data through pubsub.
func NewPubsubPublisher(ctx context.Context, host host.Host, cr routing.ContentRouting, ds datastore.Datastore, ps *pubsub.PubSub) *PubsubPublisher {
	return &PubsubPublisher{
		ctx:  ctx,
		ds:   ds,
		host: host, // needed for pubsub bootstrap
		cr:   cr,   // needed for pubsub bootstrap
		ps:   ps,
		subs: make(map[string]struct{}),
	}
}

// NewPubsubSubscriber constructs a new subscriber for arbitrary subscriptions through pubsub.
// same as above for pubsub bootstrap dependencies
func NewPubsubSubscriber(ctx context.Context, host host.Host, cr routing.ContentRouting, ds datastore.Datastore, ps *pubsub.PubSub) *PubsubSubscriber {
	return &PubsubSubscriber{
		ctx:  ctx,
		ds:   ds,
		host: host, // needed for pubsub bootstrap
		cr:   cr,   // needed for pubsub bootstrap
		ps:   ps,
		subs: make(map[string]*pubsub.Subscription),
	}
}

func (p *PubsubPublisher) Publish(ctx context.Context, topic string, data []byte) error {
	p.mx.Lock()
	id := messageTopicPrefix + topic
	_, ok := p.subs[id]

	if !ok {
		p.subs[id] = struct{}{}
		p.mx.Unlock()

		bootstrapPubsub(p.ctx, p.cr, p.host, id)
	} else {
		p.mx.Unlock()
	}

	t, err := p.ps.Join(id)
	if err != nil {
		return err
	}
	log.Debugf("PubsubPublish: publish data for %s", id)
	return t.Publish(ctx, data)
}

func (r *PubsubSubscriber) Subscribe(ctx context.Context, topic string) (chan []byte, error) {
	r.mx.Lock()
	// see if we already have a pubsub subscription; if not, subscribe
	id := messageTopicPrefix + topic
	_, ok := r.subs[id]
	resp := make(chan []byte)
	if !ok {
		t, err := r.ps.Join(id)
		if err != nil {
			return nil, err
		}
		sub, err := t.Subscribe()
		if err != nil {
			r.mx.Unlock()
			return nil, err
		}

		log.Debugf("PubsubSubscribe: subscribed to %s", id)

		r.subs[id] = sub

		ctx, cancel := context.WithCancel(r.ctx)
		go r.handleSubscription(sub, id, resp, cancel)
		go bootstrapPubsub(ctx, r.cr, r.host, id)
	}
	r.mx.Unlock()
	return resp, nil
}

// GetSubscriptions retrieves a list of active topic subscriptions
func (r *PubsubSubscriber) GetSubscriptions() []string {
	r.mx.Lock()
	defer r.mx.Unlock()

	var res []string
	for sub := range r.subs {
		res = append(res, sub)
	}

	return res
}

// Cancel cancels a topic subscription; returns true if an active
// subscription was canceled
func (r *PubsubSubscriber) Cancel(name string) bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	sub, ok := r.subs[name]
	if ok {
		sub.Cancel()
		delete(r.subs, name)
	}

	return ok
}

func (r *PubsubSubscriber) handleSubscription(sub *pubsub.Subscription, topic string, resp chan<- []byte, cancel func()) {
	defer sub.Cancel()
	defer cancel()

	for {
		msg, err := sub.Next(r.ctx)
		if err != nil {
			if err != context.Canceled {
				log.Warningf("PubsubSubscribe: subscription error in %s: %s", topic, err.Error())
			}
			return
		}

		err = r.receive(msg, topic, resp)
		if err != nil {
			log.Warningf("PubsubSubscribe: error processing update for %s: %s", topic, err.Error())
		}
	}
}

func (r *PubsubSubscriber) receive(msg *pubsub.Message, topic string, resp chan<- []byte) error {
	data := msg.GetData()
	if data == nil {
		return errors.New("empty message")
	}

	log.Debugf("PubsubSubscribe: receive data for topic %s", topic)

	resp <- data
	return nil
}

// rendezvous with peers in the name topic through provider records
// Note: rendezvous/boostrap should really be handled by the pubsub implementation itself!
func bootstrapPubsub(ctx context.Context, cr routing.ContentRouting, host host.Host, topic string) {
	topic = "floodsub:" + topic
	hash := u.Hash([]byte(topic))
	rz := cid.NewCidV1(cid.Raw, hash)

	err := cr.Provide(ctx, rz, true)
	if err != nil {
		log.Warningf("bootstrapPubsub: error providing rendezvous for %s: %s", topic, err.Error())
	}

	go func() {
		for {
			select {
			case <-time.After(8 * time.Hour):
				err := cr.Provide(ctx, rz, true)
				if err != nil {
					log.Warningf("bootstrapPubsub: error providing rendezvous for %s: %s", topic, err.Error())
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	rzctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	wg := &sync.WaitGroup{}
	for pai := range cr.FindProvidersAsync(rzctx, rz, 10) {
		if pai.ID == host.ID() {
			continue
		}
		wg.Add(1)
		go func(pai peer.AddrInfo) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			err := host.Connect(ctx, pai)
			if err != nil {
				log.Debugf("Error connecting to pubsub peer %s: %s", pai.ID, err.Error())
				return
			}

			// delay to let pubsub perform its handshake
			time.Sleep(time.Millisecond * 250)

			log.Debugf("Connected to pubsub peer %s", pai.ID)
		}(pai)
	}

	wg.Wait()
}
