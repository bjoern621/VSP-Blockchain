package discovery

import (
	"sync"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// Mock implementations for testing

type mockAddrMsgSender struct {
	mu            sync.Mutex
	sendAddrCalls []sendAddrCall
}

type sendAddrCall struct {
	peerID common.PeerId
	addrs  []PeerAddress
}

func newMockAddrMsgSender() *mockAddrMsgSender {
	return &mockAddrMsgSender{
		sendAddrCalls: make([]sendAddrCall, 0),
	}
}

func (m *mockAddrMsgSender) SendAddr(peerID common.PeerId, addrs []PeerAddress) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendAddrCalls = append(m.sendAddrCalls, sendAddrCall{
		peerID: peerID,
		addrs:  addrs,
	})
}

func (m *mockAddrMsgSender) getSendAddrCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sendAddrCalls)
}

func (m *mockAddrMsgSender) getLastSendAddrCall() *sendAddrCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sendAddrCalls) == 0 {
		return nil
	}
	return &m.sendAddrCalls[len(m.sendAddrCalls)-1]
}

type mockGetAddrMsgSender struct {
	mu               sync.Mutex
	sendGetAddrCalls []common.PeerId
	called           chan struct{}
}

func newMockGetAddrMsgSender() *mockGetAddrMsgSender {
	return &mockGetAddrMsgSender{
		sendGetAddrCalls: make([]common.PeerId, 0),
		called:           make(chan struct{}, 10),
	}
}

func (m *mockGetAddrMsgSender) SendGetAddr(peerID common.PeerId) {
	m.mu.Lock()
	m.sendGetAddrCalls = append(m.sendGetAddrCalls, peerID)
	m.mu.Unlock()
	m.called <- struct{}{}
}

func (m *mockGetAddrMsgSender) waitForCall() {
	<-m.called
}

func (m *mockGetAddrMsgSender) getSendGetAddrCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sendGetAddrCalls)
}

func (m *mockGetAddrMsgSender) getLastSendGetAddrCall() *common.PeerId {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sendGetAddrCalls) == 0 {
		return nil
	}
	return &m.sendGetAddrCalls[len(m.sendGetAddrCalls)-1]
}

type mockDiscoveryPeerRetriever struct {
	mu          sync.RWMutex
	peers       map[common.PeerId]*peer.Peer
	localPeerID common.PeerId
}

func newMockDiscoveryPeerRetriever() *mockDiscoveryPeerRetriever {
	return &mockDiscoveryPeerRetriever{
		peers:       make(map[common.PeerId]*peer.Peer),
		localPeerID: "local-peer",
	}
}

func (m *mockDiscoveryPeerRetriever) AddPeer(p *peer.Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[p.ID()] = p
}

func (m *mockDiscoveryPeerRetriever) AddPeerById(id common.PeerId, p *peer.Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[id] = p
}

func (m *mockDiscoveryPeerRetriever) GetAllPeers() []common.PeerId {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peerIds := make([]common.PeerId, 0, len(m.peers))
	for k := range m.peers {
		peerIds = append(peerIds, k)
	}
	return peerIds
}

func (m *mockDiscoveryPeerRetriever) GetAllOutboundPeers() []common.PeerId {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peerIds := make([]common.PeerId, 0, len(m.peers))
	for id, p := range m.peers {
		if p.State == common.StateConnected && p.Direction == common.DirectionOutbound {
			peerIds = append(peerIds, id)
		}
	}
	return peerIds
}

func (m *mockDiscoveryPeerRetriever) GetPeer(id common.PeerId) (*peer.Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, exists := m.peers[id]
	return p, exists
}

func (m *mockDiscoveryPeerRetriever) IsLocalPeerID(peerID common.PeerId) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return peerID == m.localPeerID
}
