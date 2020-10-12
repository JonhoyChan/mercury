package ipfs

import (
	"crypto/sha256"
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

// EncodeCID - Hash with SHA-256 and encode as a multihash
func EncodeCID(b []byte) (*cid.Cid, error) {
	multihash, err := EncodeMultihash(b)
	if err != nil {
		return nil, err
	}
	id := cid.NewCidV1(cid.Raw, *multihash)
	return &id, err
}

// EncodeMultihash - sha256 encode
func EncodeMultihash(b []byte) (*multihash.Multihash, error) {
	h := sha256.Sum256(b)
	encoded, err := multihash.Encode(h[:], multihash.SHA2_256)
	if err != nil {
		return nil, err
	}
	multihash, err := multihash.Cast(encoded)
	if err != nil {
		return nil, err
	}
	return &multihash, err
}

// ExtractIDFromPointer Certain pointers, such as moderators, contain a peerID. This function
// will extract the ID from the underlying PeerInfo object.
func ExtractIDFromPointer(pi peer.AddrInfo) (string, error) {
	if len(pi.Addrs) == 0 {
		return "", errors.New("PeerInfo object has no addresses")
	}
	addr := pi.Addrs[0]
	if addr.Protocols()[0].Code != multiaddr.P_IPFS {
		return "", errors.New("IPFS protocol not found in address")
	}
	val, err := addr.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		return "", err
	}
	return val, nil
}
