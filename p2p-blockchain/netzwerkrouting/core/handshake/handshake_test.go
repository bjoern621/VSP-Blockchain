package handshake

import (
	"os"
	"sync"
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// TestMain sets up the environment for the tests.
// For example handshake_test.go depends on ADDITIONAL_SERVICES to have a value because NewHandshakeService.InitiateHandshake() calls NewLocalVersionInfo() which depends on this environment variable. Note that ADDITIONAL_SERVICES is optional and not relevant for the tests here.
// P2P_LISTEN_ADDR is always a required environment variable.
func TestMain(m *testing.M) {
	_ = os.Setenv("P2P_LISTEN_ADDR", "does not matter")
	common.Init()
	os.Exit(m.Run())
}

type mockHandshakeMsgSender struct {
	mu           sync.Mutex
	versionCalls []common.PeerId
	verackCalls  []common.PeerId
	ackCalls     []common.PeerId
}

func newMockHandshakeMsgSender() *mockHandshakeMsgSender {
	return &mockHandshakeMsgSender{
		versionCalls: make([]common.PeerId, 0),
		verackCalls:  make([]common.PeerId, 0),
		ackCalls:     make([]common.PeerId, 0),
	}
}

func (m *mockHandshakeMsgSender) SendVersion(peerID common.PeerId, info VersionInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versionCalls = append(m.versionCalls, peerID)
}

func (m *mockHandshakeMsgSender) SendVerack(peerID common.PeerId, info VersionInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verackCalls = append(m.verackCalls, peerID)
}

func (m *mockHandshakeMsgSender) SendAck(peerID common.PeerId) {
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

	err := service.InitiateHandshake(peerID)
	if err != nil {
		t.Fatalf("unexpected error initiating handshake: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	if sender.getVersionCallCount() != 1 {
		t.Errorf("expected 1 SendVersion call, got %d", sender.getVersionCallCount())
	}

	p, ok := peerStore.GetPeer(peerID)
	if !ok {
		t.Fatal("peer should exist")
	}

	if p.State != common.StateAwaitingVerack {
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
	p.State = common.StateConnected
	p.Unlock()

	err := service.InitiateHandshake(peerID)
	if err == nil {
		t.Fatal("expected error initiating handshake for already connected peer, got nil")
	}
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

	versionInfo := VersionInfo{
		Version: "2.5.1",
	}
	versionInfo.AddService(common.ServiceType_Netzwerkrouting, common.ServiceType_BlockchainFull)

	service.HandleVersion(peerID, versionInfo)
	time.Sleep(10 * time.Millisecond)

	if sender.getVerackCallCount() != 1 {
		t.Errorf("expected 1 SendVerack call, got %d", sender.getVerackCallCount())
	}

	p, _ := peerStore.GetPeer(peerID)

	if p.State != common.StateAwaitingAck {
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
	p.State = common.StateAwaitingVerack
	p.Unlock()

	versionInfo := VersionInfo{
		Version: "1.5.0",
	}
	versionInfo.AddService(common.ServiceType_BlockchainFull, common.ServiceType_Netzwerkrouting, common.ServiceType_Miner)

	service.HandleVerack(peerID, versionInfo)
	time.Sleep(10 * time.Millisecond)

	if sender.getAckCallCount() != 1 {
		t.Errorf("expected 1 SendAck call, got %d", sender.getAckCallCount())
	}

	if p.State != common.StateConnected {
		t.Errorf("expected state StateConnected, got %v", p.State)
	}
	if p.Version != "1.5.0" {
		t.Errorf("expected version 1.5.0, got %s", p.Version)
	}
	if len(p.SupportedServices) != 3 {
		t.Errorf("expected 3 supported service, got %d", len(p.SupportedServices))
	}
}

func TestHandleAck(t *testing.T) {
	peerStore := peer.NewPeerStore()
	sender := newMockHandshakeMsgSender()
	service := NewHandshakeService(sender, peerStore)

	peerID := peerStore.NewInboundPeer()

	p, _ := peerStore.GetPeer(peerID)
	p.Lock()
	p.State = common.StateAwaitingAck
	p.Unlock()

	service.HandleAck(peerID)

	if p.State != common.StateConnected {
		t.Errorf("expected state StateConnected, got %v", p.State)
	}
}
