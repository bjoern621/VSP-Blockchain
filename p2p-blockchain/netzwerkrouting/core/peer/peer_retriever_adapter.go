// Package peer provides adapters for peer-related interfaces.
package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

// PeerData is an interface for accessing peer information.
type PeerData interface {
	Lock()
	Unlock()
	GetVersion() string
	GetState() common.PeerConnectionState
	GetSupportedServices() []common.ServiceType
	GetLastSeen() int64
}

// dataPeerRetriever is the interface that the data layer's PeerStore implements.
type dataPeerRetriever interface {
	GetPeer(id common.PeerId) (*common.Peer, bool)
}

// peerDataAdapter wraps data layer's Peer and implements PeerData interface.
type peerDataAdapter struct {
	peer *common.Peer
}

func (a *peerDataAdapter) Lock() {
	a.peer.Lock()
}

func (a *peerDataAdapter) Unlock() {
	a.peer.Unlock()
}

func (a *peerDataAdapter) GetVersion() string {
	return a.peer.Version
}

func (a *peerDataAdapter) GetState() common.PeerConnectionState {
	return a.peer.State
}

func (a *peerDataAdapter) GetSupportedServices() []common.ServiceType {
	return a.peer.SupportedServices
}

func (a *peerDataAdapter) GetLastSeen() int64 {
	return a.peer.LastSeen
}

// PeerRetrieverAdapter adapts the data layer's peer store to the API layer's PeerRetriever interface.
type PeerRetrieverAdapter struct {
	peerStore dataPeerRetriever
}

// NewPeerRetrieverAdapter creates a new adapter wrapping the given peer store.
func NewPeerRetrieverAdapter(peerStore dataPeerRetriever) *PeerRetrieverAdapter {
	return &PeerRetrieverAdapter{peerStore: peerStore}
}

// GetPeer retrieves a peer by ID and returns it as PeerData.
func (a *PeerRetrieverAdapter) GetPeer(id common.PeerId) (PeerData, bool) {
	p, exists := a.peerStore.GetPeer(id)
	if !exists {
		return nil, false
	}
	return &peerDataAdapter{peer: p}, true
}
