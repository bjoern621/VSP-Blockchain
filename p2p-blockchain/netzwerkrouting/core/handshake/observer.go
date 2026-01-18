package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	mapset "github.com/deckarep/golang-set/v2"
)

// ConnectionObserver defines the interface for observing connection state changes.
// Implementations can react to peer connections being established.
type ConnectionObserver interface {
	// OnPeerConnected is called when a peer's handshake completes and the connection is fully established.
	// isOutbound indicates if this node initiated the connection (true) or received it (false).
	// This is the trigger point for initiating the Initial Block Download (IBD) for outbound connections.
	OnPeerConnected(peerID common.PeerId, isOutbound bool)
}

// ObservableHandshakeService extends the handshake service with observer pattern support.
type ObservableHandshakeService interface {
	// Attach registers a ConnectionObserver to receive connection notifications.
	Attach(o ConnectionObserver)
	// Detach removes a previously registered ConnectionObserver.
	Detach(o ConnectionObserver)
}

// observerManager handles the registration and notification of connection observers.
type observerManager struct {
	observers mapset.Set[ConnectionObserver]
}

func newObserverManager() *observerManager {
	return &observerManager{
		observers: mapset.NewSet[ConnectionObserver](),
	}
}

func (m *observerManager) Attach(o ConnectionObserver) {
	m.observers.Add(o)
}

func (m *observerManager) Detach(o ConnectionObserver) {
	m.observers.Remove(o)
}

func (m *observerManager) notifyPeerConnected(peerID common.PeerId, isOutbound bool) {
	for o := range m.observers.Iter() {
		go o.OnPeerConnected(peerID, isOutbound)
	}
}
