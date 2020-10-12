package ipfs

import (
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/ipfs/go-ipfs/core"
)

func ConnectedPeers(n *core.IpfsNode) []peer.ID {
	return n.PeerHost.Network().Peers()
}
