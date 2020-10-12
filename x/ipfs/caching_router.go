package ipfs

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p-kad-dht"
)

var (
	ErrCachingRouterIncorrectRoutingType = errors.New("Incorrect routing type")
)

type CachingRouter struct {
	apiRouter *APIRouter
	routing.Routing
}

func NewCachingRouter(dht *dht.IpfsDHT, apiRouter *APIRouter) *CachingRouter {
	return &CachingRouter{
		apiRouter: apiRouter,
		Routing:   dht,
	}
}

func (r *CachingRouter) DHT() (*dht.IpfsDHT, error) {
	dht, ok := r.Routing.(*dht.IpfsDHT)
	if !ok {
		return nil, ErrCachingRouterIncorrectRoutingType
	}
	return dht, nil
}

func (r *CachingRouter) APIRouter() *APIRouter {
	return r.apiRouter
}

func (r *CachingRouter) PutValue(ctx context.Context, key string, value []byte, opts ...routing.Option) error {
	// Write to the tiered router in the background then write to the caching
	// router and return
	var err error
	if err = r.Routing.PutValue(ctx, key, value, opts...); err != nil {
		log.Errorf("ipfs dht put (%s): %s", hex.EncodeToString([]byte(key)), err)
		return err
	}
	if err = r.apiRouter.PutValue(ctx, key, value, opts...); err != nil {
		log.Errorf("api cache put (%s): %s", hex.EncodeToString([]byte(key)), err)
	}
	return err
}

func (r *CachingRouter) GetValue(ctx context.Context, key string, opts ...routing.Option) ([]byte, error) {
	// First check the DHT router. If it's successful return the value otherwise
	// continue on to check the other routers.
	val, err := r.Routing.GetValue(ctx, key, opts...)
	if err != nil && len(val) == 0 {
		// No values from the DHT, check the API cache
		log.Warningf("ipfs dht lookup was empty: %s", err.Error())
		if val, err = r.apiRouter.GetValue(ctx, key, opts...); err != nil && len(val) == 0 {
			// No values still, report NotFound
			return nil, routing.ErrNotFound
		}
	}
	if err := r.apiRouter.PutValue(ctx, key, val, opts...); err != nil {
		log.Errorf("api cache put found dht value (%s): %s", hex.EncodeToString([]byte(key)), err.Error())
	}
	return val, nil
}

func (r *CachingRouter) GetPublicKey(ctx context.Context, p peer.ID) (crypto.PubKey, error) {
	if dht, ok := r.Routing.(routing.PubKeyFetcher); ok {
		return dht.GetPublicKey(ctx, p)
	}
	return nil, routing.ErrNotSupported
}

func (r *CachingRouter) SearchValue(ctx context.Context, key string, opts ...routing.Option) (<-chan []byte, error) {
	// TODO: Restore parallel lookup once validation is properly applied to
	// the apiRouter results ensuring it doesn't return invalid records before the
	// IpfsRouting object can. For some reason the validation is not being considered
	// on returned results.
	return r.Routing.SearchValue(ctx, key, opts...)
	//return routinghelpers.Parallel{
	//Routers: []routing.IpfsRouting{
	//r.IpfsRouting,
	//r.apiRouter,
	//},
	//Validator: r.RecordValidator,
	//}.SearchValue(ctx, key, opts...)
}
