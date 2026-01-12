package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// PeerCreator is an interface for creating new peers.
// It is defined here in the business layer only because api layer needs it.
// It is identical to the one in data layer.
//
// Note that there is no implementation of this interface in the business layer.
// Instead, the data layer's PeerStore implements it.
// This is not a problem, because the api layer can depend on the business layer statically through this interface and only needs data layer at runtime.
type PeerCreator interface {
	peer.PeerCreator
}
