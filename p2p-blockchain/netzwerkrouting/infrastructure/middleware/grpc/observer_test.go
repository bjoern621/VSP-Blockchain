package grpc

import (
	"reflect"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/block"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"testing"
	"time"
)

// MockObserver struct used to verify that the Server correctly notifies observers.
type MockObserver struct {
	InvCh         chan block.InvMsg
	GetDataCh     chan block.GetDataMsg
	BlockCh       chan block.BlockMsg
	MerkleBlockCh chan block.MerkleBlockMsg
	TxCh          chan block.TxMsg
	GetHeadersCh  chan block.BlockLocator
	HeadersCh     chan []block.BlockHeader
	SetFilterCh   chan block.SetFilterRequest
	MempoolCh     chan struct{}
}

// NewMockObserver creates a MockObserver with buffered channels to prevent blocking during tests.
func NewMockObserver() *MockObserver {
	return &MockObserver{
		InvCh:         make(chan block.InvMsg, 10),
		GetDataCh:     make(chan block.GetDataMsg, 10),
		BlockCh:       make(chan block.BlockMsg, 10),
		MerkleBlockCh: make(chan block.MerkleBlockMsg, 10),
		TxCh:          make(chan block.TxMsg, 10),
		GetHeadersCh:  make(chan block.BlockLocator, 10),
		HeadersCh:     make(chan []block.BlockHeader, 10),
		SetFilterCh:   make(chan block.SetFilterRequest, 10),
		MempoolCh:     make(chan struct{}, 10),
	}
}

// Implement the BlockchainObserverAPI interface

func (m *MockObserver) Inv(invMsg block.InvMsg, _ peer.PeerID) {
	m.InvCh <- invMsg
}

func (m *MockObserver) GetData(getDataMsg block.GetDataMsg, _ peer.PeerID) {
	m.GetDataCh <- getDataMsg
}

func (m *MockObserver) Block(blockMsg block.BlockMsg, _ peer.PeerID) {
	m.BlockCh <- blockMsg
}

func (m *MockObserver) MerkleBlock(merkleBlockMsg block.MerkleBlockMsg, _ peer.PeerID) {
	m.MerkleBlockCh <- merkleBlockMsg
}

func (m *MockObserver) Tx(txMsg block.TxMsg, _ peer.PeerID) {
	m.TxCh <- txMsg
}

func (m *MockObserver) GetHeaders(locator block.BlockLocator, _ peer.PeerID) {
	m.GetHeadersCh <- locator
}

func (m *MockObserver) Headers(headers []block.BlockHeader, _ peer.PeerID) {
	m.HeadersCh <- headers
}

func (m *MockObserver) SetFilter(setFilterRequest block.SetFilterRequest, _ peer.PeerID) {
	m.SetFilterCh <- setFilterRequest
}

func (m *MockObserver) Mempool(_ peer.PeerID) {
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
		msg := block.InvMsg{
			Inventory: []block.InvVector{{InvType: 1}},
		}

		server.NotifyInv(msg, testPeerID)

		select {
		case received := <-mockObs.InvCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyInv")
		}
	})

	t.Run("NotifyGetData", func(t *testing.T) {
		msg := block.GetDataMsg{
			Inventory: []block.InvVector{{InvType: 1}},
		}

		server.NotifyGetData(msg, testPeerID)

		select {
		case received := <-mockObs.GetDataCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyGetData")
		}
	})

	t.Run("NotifyBlock", func(t *testing.T) {
		msg := block.BlockMsg{
			Block: block.Block{Header: block.BlockHeader{Nonce: 123}},
		}

		server.NotifyBlock(msg, testPeerID)

		select {
		case received := <-mockObs.BlockCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyBlock")
		}
	})

	t.Run("NotifyMerkleBlock", func(t *testing.T) {
		msg := block.MerkleBlockMsg{
			MerkleBlock: block.MerkleBlock{
				BlockHeader: block.BlockHeader{Nonce: 456},
				Proofs:      []block.MerkleProof{{Index: 1}},
			},
		}

		server.NotifyMerkleBlock(msg, testPeerID)

		select {
		case received := <-mockObs.MerkleBlockCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyMerkleBlock")
		}
	})

	t.Run("NotifyTx", func(t *testing.T) {
		msg := block.TxMsg{
			Transaction: transaction.Transaction{LockTime: 500},
		}

		server.NotifyTx(msg, testPeerID)

		select {
		case received := <-mockObs.TxCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyTx")
		}
	})

	t.Run("NotifyGetHeaders", func(t *testing.T) {
		msg := block.BlockLocator{
			// Add specific fields if necessary for meaningful test
		}

		server.NotifyGetHeaders(msg, testPeerID)

		select {
		case received := <-mockObs.GetHeadersCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyGetHeaders")
		}
	})

	t.Run("NotifyHeaders", func(t *testing.T) {
		msg := []block.BlockHeader{
			{Nonce: 789},
			{Nonce: 101},
		}

		server.NotifyHeaders(msg, testPeerID)

		select {
		case received := <-mockObs.HeadersCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyHeaders")
		}
	})

	t.Run("NotifySetFilterRequest", func(t *testing.T) {
		msg := block.SetFilterRequest{
			// Add specific fields if necessary
		}

		server.NotifySetFilterRequest(msg, testPeerID)

		select {
		case received := <-mockObs.SetFilterCh:
			if !reflect.DeepEqual(received, msg) {
				t.Errorf("Expected message %v, got %v", msg, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifySetFilterRequest")
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
