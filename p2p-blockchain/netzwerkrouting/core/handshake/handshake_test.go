package handshake

import (
	"os"
	"sync"
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// TestMain sets up the environment for the tests.
// For example handshake_test.go depends on ADDITIONAL_SERVICES to have a value because NewHandshakeService.InitiateHandshake() calls NewLocalVersionInfo() which depends on this environment variable. Note that ADDITIONAL_SERVICES is optional and not relevant for the tests here.
// P2P_LISTEN_ADDR is always a required environment variable.
func TestMain(m *testing.M) {
	os.Setenv("P2P_LISTEN_ADDR", "does not matter")
	common.Init()
	os.Exit(m.Run())
}

type mockHandshakeMsgSender struct {
	mu           sync.Mutex
	versionCalls []peer.PeerID
	verackCalls  []peer.PeerID
	ackCalls     []peer.PeerID
}

func newMockHandshakeMsgSender() *mockHandshakeMsgSender {
	return &mockHandshakeMsgSender{
		versionCalls: make([]peer.PeerID, 0),
		verackCalls:  make([]peer.PeerID, 0),
		ackCalls:     make([]peer.PeerID, 0),
	}
}

func (m *mockHandshakeMsgSender) SendVersion(peerID peer.PeerID, info VersionInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versionCalls = append(m.versionCalls, peerID)
}

func (m *mockHandshakeMsgSender) SendVerack(peerID peer.PeerID, info VersionInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verackCalls = append(m.verackCalls, peerID)
}

func (m *mockHandshakeMsgSender) SendAck(peerID peer.PeerID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ackCalls = append(m.ackCalls, peerID)
}

func (m *mockHandshakeMsgSender) getVersionCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.versionCalls)
}

func (m *mockHandshakeMsgSender) getVerackCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.verackCalls)
}

func (m *mockHandshakeMsgSender) getAckCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.ackCalls)
}

func TestInitiateHandshake(t *testing.T) {
	peerStore := peer.NewPeerStore()
	sender := newMockHandshakeMsgSender()
	service := NewHandshakeService(sender, peerStore)

	peerID := peerStore.NewOutboundPeer()

	service.InitiateHandshake(peerID)
	time.Sleep(10 * time.Millisecond)

	if sender.getVersionCallCount() != 1 {
		t.Errorf("expected 1 SendVersion call, got %d", sender.getVersionCallCount())
	}

	p, ok := peerStore.GetPeer(peerID)
	if !ok {
		t.Fatal("peer should exist")
	}

	if p.State != peer.StateAwaitingVerack {
		t.Errorf("expected state StateAwaitingVerack, got %v", p.State)
	}
}

func TestInitiateHandshake_RejectsWhenAlreadyConnected(t *testing.T) {
	peerStore := peer.NewPeerStore()
	sender := newMockHandshakeMsgSender()
	service := NewHandshakeService(sender, peerStore)

	peerID := peerStore.NewOutboundPeer()

	p, _ := peerStore.GetPeer(peerID)
	p.Lock()
	p.State = peer.StateConnected
	p.Unlock()

	service.InitiateHandshake(peerID)
	time.Sleep(10 * time.Millisecond)

	if sender.getVersionCallCount() != 0 {
		t.Errorf("expected 0 SendVersion calls for already connected peer, got %d", sender.getVersionCallCount())
	}
}

func TestHandleVersion(t *testing.T) {
	peerStore := peer.NewPeerStore()
	sender := newMockHandshakeMsgSender()
	service := NewHandshakeService(sender, peerStore)

	peerID := peerStore.NewInboundPeer()

	services := []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull}
	versionInfo := VersionInfo{
		Version:           "2.5.1",
		SupportedServices: services,
	}

	service.HandleVersion(peerID, versionInfo)
	time.Sleep(10 * time.Millisecond)

	if sender.getVerackCallCount() != 1 {
		t.Errorf("expected 1 SendVerack call, got %d", sender.getVerackCallCount())
	}

	p, _ := peerStore.GetPeer(peerID)

	if p.State != peer.StateAwaitingAck {
		t.Errorf("expected state StateAwaitingAck, got %v", p.State)
	}
	if p.Version != "2.5.1" {
		t.Errorf("expected version 2.5.1, got %s", p.Version)
	}
	if len(p.SupportedServices) != 2 {
		t.Errorf("expected 2 supported services, got %d", len(p.SupportedServices))
	}
}

func TestHandleVerack(t *testing.T) {
	peerStore := peer.NewPeerStore()
	sender := newMockHandshakeMsgSender()
	service := NewHandshakeService(sender, peerStore)

	peerID := peerStore.NewOutboundPeer()

	p, _ := peerStore.GetPeer(peerID)
	p.Lock()
	p.State = peer.StateAwaitingVerack
	p.Unlock()

	services := []peer.ServiceType{peer.ServiceType_Miner}
	versionInfo := VersionInfo{
		Version:           "1.5.0",
		SupportedServices: services,
	}

	service.HandleVerack(peerID, versionInfo)
	time.Sleep(10 * time.Millisecond)

	if sender.getAckCallCount() != 1 {
		t.Errorf("expected 1 SendAck call, got %d", sender.getAckCallCount())
	}

	if p.State != peer.StateConnected {
		t.Errorf("expected state StateConnected, got %v", p.State)
	}
	if p.Version != "1.5.0" {
		t.Errorf("expected version 1.5.0, got %s", p.Version)
	}
	if len(p.SupportedServices) != 1 {
		t.Errorf("expected 1 supported service, got %d", len(p.SupportedServices))
	}
}

func TestHandleAck(t *testing.T) {
	peerStore := peer.NewPeerStore()
	sender := newMockHandshakeMsgSender()
	service := NewHandshakeService(sender, peerStore)

	peerID := peerStore.NewInboundPeer()

	p, _ := peerStore.GetPeer(peerID)
	p.Lock()
	p.State = peer.StateAwaitingAck
	p.Unlock()

	service.HandleAck(peerID)

	if p.State != peer.StateConnected {
		t.Errorf("expected state StateConnected, got %v", p.State)
	}
}
