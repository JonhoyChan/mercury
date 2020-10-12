package ipfs

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
)

// Query returns the closest peers known for peerID
func Query(dht *dht.IpfsDHT, peerID string) ([]peer.ID, error) {
	id, err := peer.Decode(peerID)
	if err != nil {
		return nil, err
	}
	ch, err := dht.GetClosestPeers(context.Background(), string(id))
	if err != nil {
		return nil, err
	}

	var closestPeers []peer.ID
	for p := range ch {
		closestPeers = append(closestPeers, p)
	}
	return closestPeers, nil
}
