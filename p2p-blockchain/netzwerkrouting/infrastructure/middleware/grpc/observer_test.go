package grpc

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"testing"
	"time"
)

// MockObserver struct used to verify that the Server correctly notifies observers.
type MockObserver struct {
	InvCh         chan *blockchain.InvMsg
	GetDataCh     chan *blockchain.GetDataMsg
	BlockCh       chan *blockchain.BlockMsg
	MerkleBlockCh chan *blockchain.MerkleBlockMsg
	TxCh          chan *blockchain.TxMsg
	GetHeadersCh  chan *blockchain.BlockLocator
	HeadersCh     chan []*blockchain.BlockHeader
	SetFilterCh   chan *blockchain.SetFilterRequest
	MempoolCh     chan struct{}
}

// NewMockObserver creates a MockObserver with buffered channels to prevent blocking during tests.
func NewMockObserver() *MockObserver {
	return &MockObserver{
		InvCh:         make(chan *blockchain.InvMsg, 10),
		GetDataCh:     make(chan *blockchain.GetDataMsg, 10),
		BlockCh:       make(chan *blockchain.BlockMsg, 10),
		MerkleBlockCh: make(chan *blockchain.MerkleBlockMsg, 10),
		TxCh:          make(chan *blockchain.TxMsg, 10),
		GetHeadersCh:  make(chan *blockchain.BlockLocator, 10),
		HeadersCh:     make(chan []*blockchain.BlockHeader, 10),
		SetFilterCh:   make(chan *blockchain.SetFilterRequest, 10),
		MempoolCh:     make(chan struct{}, 10),
	}
}

// Implement the BlockchainObserverAPI interface

func (m *MockObserver) Inv(invMsg *blockchain.InvMsg, peerID peer.PeerID) {
	m.InvCh <- invMsg
}

func (m *MockObserver) GetData(getDataMsg *blockchain.GetDataMsg, peerID peer.PeerID) {
	m.GetDataCh <- getDataMsg
}

func (m *MockObserver) Block(blockMsg *blockchain.BlockMsg, peerID peer.PeerID) {
	m.BlockCh <- blockMsg
}

func (m *MockObserver) MerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg, peerID peer.PeerID) {
	m.MerkleBlockCh <- merkleBlockMsg
}

func (m *MockObserver) Tx(txMsg *blockchain.TxMsg, peerID peer.PeerID) {
	m.TxCh <- txMsg
}

func (m *MockObserver) GetHeaders(locator *blockchain.BlockLocator, peerID peer.PeerID) {
	m.GetHeadersCh <- locator
}

func (m *MockObserver) Headers(headers []*blockchain.BlockHeader, peerID peer.PeerID) {
	m.HeadersCh <- headers
}

func (m *MockObserver) SetFilter(setFilterRequest *blockchain.SetFilterRequest, peerID peer.PeerID) {
	m.SetFilterCh <- setFilterRequest
}

func (m *MockObserver) Mempool(peerID peer.PeerID) {
	m.MempoolCh <- struct{}{}
}

// Tests

func TestObserverBlockchainServer_Notify(t *testing.T) {
	// 1. Create the Mock Observer
	mockObs := NewMockObserver()

	// 2. Instantiate the Server manually.
	// We avoid NewServer() or gRPC logic because we only want to test the observer pattern logic.
	// We manually initialize the 'observers' map since we are in the same package (grpc).
	server := &Server{
		observers: make(map[observer.BlockchainObserverAPI]struct{}),
	}

	// 3. Attach the mock observer
	// We can use server.Attach(mockObs) if available, or inject directly:
	server.observers[mockObs] = struct{}{}

	// Shared test data
	testPeerID := peer.PeerID("test-peer-123")
	timeout := time.Millisecond * 100

	t.Run("NotifyInv", func(t *testing.T) {
		msg := &blockchain.InvMsg{
			Inventory: []*blockchain.InvVector{{InvType: 1}},
		}

		// Act
		server.NotifyInv(msg, testPeerID)

		// Assert
		select {
		case received := <-mockObs.InvCh:
			if received != msg {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyInv")
		}
	})

	t.Run("NotifyGetData", func(t *testing.T) {
		msg := &blockchain.GetDataMsg{
			Inventory: []*blockchain.InvVector{{InvType: 1}},
		}

		server.NotifyGetData(msg, testPeerID)

		select {
		case received := <-mockObs.GetDataCh:
			if received != msg {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyGetData")
		}
	})

	t.Run("NotifyBlock", func(t *testing.T) {
		msg := &blockchain.BlockMsg{
			Block: &blockchain.Block{Header: &blockchain.BlockHeader{Nonce: 123}},
		}

		server.NotifyBlock(msg, testPeerID)

		select {
		case received := <-mockObs.BlockCh:
			if received != msg {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyBlock")
		}
	})

	t.Run("NotifyTx", func(t *testing.T) {
		msg := &blockchain.TxMsg{
			Transaction: &blockchain.Transaction{LockTime: 500},
		}

		server.NotifyTx(msg, testPeerID)

		select {
		case received := <-mockObs.TxCh:
			if received != msg {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyTx")
		}
	})

	t.Run("NotifyMempool", func(t *testing.T) {
		server.NotifyMempool(testPeerID)

		select {
		case <-mockObs.MempoolCh:
			// Success
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyMempool")
		}
	})
}
