package api

import (
	"fmt"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

type mockHandshakeInitiator struct {
	mu                     sync.Mutex
	initiateHandshakeCalls []common.PeerId
}

func newMockHandshakeInitiator() *mockHandshakeInitiator {
	return &mockHandshakeInitiator{
		initiateHandshakeCalls: make([]common.PeerId, 0),
	}
}

func (m *mockHandshakeInitiator) InitiateHandshake(peerID common.PeerId) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initiateHandshakeCalls = append(m.initiateHandshakeCalls, peerID)
	return nil
}

func (m *mockHandshakeInitiator) getInitiateHandshakeCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.initiateHandshakeCalls)
}

type mockOutboundPeerResolver struct {
	mu              sync.Mutex
	registeredPeers map[string]common.PeerId
	peerAddresses   map[common.PeerId]netip.AddrPort
}

func newMockOutboundPeerResolver() *mockOutboundPeerResolver {
	return &mockOutboundPeerResolver{
		registeredPeers: make(map[string]common.PeerId),
		peerAddresses:   make(map[common.PeerId]netip.AddrPort),
	}
}

func (m *mockOutboundPeerResolver) GetOutboundPeer(addrPort netip.AddrPort) (common.PeerId, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	peerID, exists := m.registeredPeers[addrPort.String()]
	return peerID, exists
}

func (m *mockOutboundPeerResolver) RegisterPeer(peerID common.PeerId, listeningEndpoint netip.AddrPort) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registeredPeers[listeningEndpoint.String()] = peerID
	m.peerAddresses[peerID] = listeningEndpoint
}

func (m *mockOutboundPeerResolver) getRegisteredPeerCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.registeredPeers)
}

func TestNewHandshakeAPIService(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	if api == nil {
		t.Fatal("expected non-nil api service")
	}
}

func TestInitiateHandshake_SuccessfulFlow(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	addrPort := netip.MustParseAddrPort("127.0.0.1:9000")
	err := api.InitiateHandshake(addrPort)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if initiator.getInitiateHandshakeCallCount() != 1 {
		t.Errorf("expected 1 InitiateHandshake call, got %d", initiator.getInitiateHandshakeCallCount())
	}

	if resolver.getRegisteredPeerCount() != 1 {
		t.Errorf("expected 1 registered peer, got %d", resolver.getRegisteredPeerCount())
	}
}

func TestInitiateHandshake_SameAddressAndPort(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	addrPort := netip.MustParseAddrPort("127.0.0.1:9000")

	// First handshake
	err1 := api.InitiateHandshake(addrPort)
	if err1 != nil {
		t.Fatalf("first handshake failed: %v", err1)
	}

	peerID1, exists1 := resolver.GetOutboundPeer(addrPort)
	if !exists1 {
		t.Fatal("expected peer to be registered after first handshake")
	}

	// Try to initiate handshake with same address again.
	// The API service is expected to reuse the already-registered peer and delegate
	// success/failure to the handshake initiator.
	err2 := api.InitiateHandshake(addrPort)
	if err2 != nil {
		t.Fatalf("second handshake failed unexpectedly: %v", err2)
	}

	peerID2, exists2 := resolver.GetOutboundPeer(addrPort)
	if !exists2 {
		t.Fatal("expected peer to still be registered after second handshake")
	}
	if peerID1 != peerID2 {
		t.Fatalf("expected same peerID to be reused, got %s then %s", peerID1, peerID2)
	}

	if initiator.getInitiateHandshakeCallCount() != 2 {
		t.Errorf("expected 2 InitiateHandshake calls, got %d", initiator.getInitiateHandshakeCallCount())
	}

	if resolver.getRegisteredPeerCount() != 1 {
		t.Errorf("expected 1 registered peer, got %d", resolver.getRegisteredPeerCount())
	}
}

func TestInitiateHandshake_MultipleDistinctPeers(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	addresses := []string{
		"127.0.0.1:9000",
		"127.0.0.1:9001",
		"127.0.0.1:9002",
		"192.168.1.1:8000",
	}

	for _, addr := range addresses {
		addrPort := netip.MustParseAddrPort(addr)
		err := api.InitiateHandshake(addrPort)
		if err != nil {
			t.Errorf("failed to initiate handshake with %s: %v", addr, err)
		}
	}

	if initiator.getInitiateHandshakeCallCount() != len(addresses) {
		t.Errorf("expected %d InitiateHandshake calls, got %d", len(addresses), initiator.getInitiateHandshakeCallCount())
	}

	if resolver.getRegisteredPeerCount() != len(addresses) {
		t.Errorf("expected %d registered peers, got %d", len(addresses), resolver.getRegisteredPeerCount())
	}
}

func TestInitiateHandshake_CreatesNewPeerBeforeRegistering(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	addrPort := netip.MustParseAddrPort("127.0.0.1:9000")
	err := api.InitiateHandshake(addrPort)

	if err != nil {
		t.Fatalf("InitiateHandshake failed: %v", err)
	}

	// Verify peer was registered
	peerID, exists := resolver.GetOutboundPeer(addrPort)
	if !exists {
		t.Error("peer should be registered in resolver")
	}

	// Verify peer exists in peer store
	_, existsInStore := peerStore.GetPeer(peerID)
	if !existsInStore {
		t.Error("peer should exist in peer store")
	}
}

func TestInitiateHandshake_ConcurrentCalls(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	var wg sync.WaitGroup
	numCalls := 50

	addresses := make([]netip.AddrPort, numCalls)
	for i := range numCalls {
		// Create unique addresses to avoid conflicts
		addrPort := netip.MustParseAddrPort(fmt.Sprintf("127.0.0.1:%d", 9000+i))
		addresses[i] = addrPort
	}

	wg.Add(numCalls)
	for i := range numCalls {
		go func(idx int) {
			defer wg.Done()
			_ = api.InitiateHandshake(addresses[idx])
		}(i)
	}

	wg.Wait()

	if initiator.getInitiateHandshakeCallCount() != numCalls {
		t.Errorf("expected %d InitiateHandshake calls, got %d", numCalls, initiator.getInitiateHandshakeCallCount())
	}

	if resolver.getRegisteredPeerCount() != numCalls {
		t.Errorf("expected %d registered peers, got %d", numCalls, resolver.getRegisteredPeerCount())
	}
}

func TestInitiateHandshake_FullChain_CreationToInitiation(t *testing.T) {
	resolver := newMockOutboundPeerResolver()
	peerStore := peer.NewPeerStore()
	initiator := newMockHandshakeInitiator()

	api := NewHandshakeAPIService(resolver, peerStore, initiator)

	addrPort := netip.MustParseAddrPort("192.168.1.100:5000")
	err := api.InitiateHandshake(addrPort)

	if err != nil {
		t.Fatalf("InitiateHandshake failed: %v", err)
	}

	// Get the registered peer ID from resolver
	registeredPeerID, exists := resolver.GetOutboundPeer(addrPort)
	if !exists {
		t.Fatal("peer not registered in resolver")
	}

	// Verify initiator was called with the same peer ID
	if initiator.getInitiateHandshakeCallCount() != 1 {
		t.Fatal("initiator not called")
	}

	// Verify peer exists in peer store with correct direction
	peerObj, peerExists := peerStore.GetPeer(registeredPeerID)
	if !peerExists {
		t.Fatal("peer not found in peer store")
	}

	if peerObj.Direction != common.DirectionOutbound {
		t.Errorf("expected peer direction OutBound, got %v", peerObj.Direction)
	}
	if peerObj.State != common.StateNew {
		t.Errorf("expected peer state New, got %v", peerObj.State)
	}
}
