package ipfs

import (
	"context"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-record"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo"
)

// PrepareIPFSConfig builds the configuration options for the internal IPFS node.
func PrepareIPFSConfig(r repo.Repo) *core.BuildCfg {
	ncfg := &core.BuildCfg{
		Repo:      r,
		Online:    true,
		ExtraOpts: map[string]bool{},
	}
	ncfg.Routing = constructRouting
	return ncfg
}

func constructRouting(ctx context.Context, host host.Host, dstore datastore.Batching, validator record.Validator, peers ...peer.AddrInfo) (routing.Routing, error) {
	return dht.New(
		ctx, host,
		dht.Datastore(dstore),
		dht.Validator(validator),
		dht.ProtocolPrefix("mercury"),
	)
}
