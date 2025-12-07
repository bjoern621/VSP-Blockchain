package grpc

import (
	"context"
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/observer"
)

// MockObserver implements observer.BlockchainObserver for testing purposes.
// It uses channels to signal when methods are called, allowing synchronization in tests.
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

func NewMockObserver() *MockObserver {
	return &MockObserver{
		InvCh:         make(chan *blockchain.InvMsg, 1),
		GetDataCh:     make(chan *blockchain.GetDataMsg, 1),
		BlockCh:       make(chan *blockchain.BlockMsg, 1),
		MerkleBlockCh: make(chan *blockchain.MerkleBlockMsg, 1),
		TxCh:          make(chan *blockchain.TxMsg, 1),
		GetHeadersCh:  make(chan *blockchain.BlockLocator, 1),
		HeadersCh:     make(chan []*blockchain.BlockHeader, 1),
		SetFilterCh:   make(chan *blockchain.SetFilterRequest, 1),
		MempoolCh:     make(chan struct{}, 1),
	}
}

func (m *MockObserver) Inv(msg *blockchain.InvMsg) {
	m.InvCh <- msg
}
func (m *MockObserver) GetData(msg *blockchain.GetDataMsg) {
	m.GetDataCh <- msg
}
func (m *MockObserver) Block(msg *blockchain.BlockMsg) {
	m.BlockCh <- msg
}
func (m *MockObserver) MerkleBlock(msg *blockchain.MerkleBlockMsg) {
	m.MerkleBlockCh <- msg
}
func (m *MockObserver) Tx(msg *blockchain.TxMsg) {
	m.TxCh <- msg
}
func (m *MockObserver) GetHeaders(locator *blockchain.BlockLocator) {
	m.GetHeadersCh <- locator
}
func (m *MockObserver) Headers(headers []*blockchain.BlockHeader) {
	m.HeadersCh <- headers
}
func (m *MockObserver) SetFilter(request *blockchain.SetFilterRequest) {
	m.SetFilterCh <- request
}
func (m *MockObserver) Mempool() {
	m.MempoolCh <- struct{}{}
}

// Ensure MockObserver implements the interface
var _ observer.BlockchainObserver = (*MockObserver)(nil)

func setupTestServer() (*Server, *MockObserver) {
	// Pass nil for dependencies as we are testing the observer pattern wiring only.
	// NewServer implementation should handle nil if it doesn't use them in the constructor.
	server := NewServer(nil, nil)
	mockObs := NewMockObserver()
	server.Attach(mockObs)
	return server, mockObs
}

func TestServer_ObserverPattern(t *testing.T) {
	// Helper to create 32-byte hash required by the converters
	dummyHash := make([]byte, 32)
	for i := range dummyHash {
		dummyHash[i] = 0xAA
	}

	t.Run("Inv", func(t *testing.T) {
		server, mockObs := setupTestServer()
		invMsg := &pb.InvMsg{
			Inventory: []*pb.InvVector{
				{Type: 1, Hash: dummyHash},
			},
		}

		_, err := server.Inv(context.Background(), invMsg)
		if err != nil {
			t.Fatalf("Inv failed: %v", err)
		}

		select {
		case msg := <-mockObs.InvCh:
			if len(msg.Inventory) != 1 {
				t.Errorf("Expected 1 inventory item, got %d", len(msg.Inventory))
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for Inv notification")
		}
	})

	t.Run("GetData", func(t *testing.T) {
		server, mockObs := setupTestServer()
		getDataMsg := &pb.GetDataMsg{
			Inventory: []*pb.InvVector{
				{Type: 1, Hash: dummyHash},
			},
		}

		_, err := server.GetData(context.Background(), getDataMsg)
		if err != nil {
			t.Fatalf("GetData failed: %v", err)
		}

		select {
		case msg := <-mockObs.GetDataCh:
			if len(msg.Inventory) != 1 {
				t.Errorf("Expected 1 inventory item, got %d", len(msg.Inventory))
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for GetData notification")
		}
	})

	t.Run("Block", func(t *testing.T) {
		server, mockObs := setupTestServer()
		blockMsg := &pb.BlockMsg{
			Block: &pb.Block{
				Header: &pb.BlockHeader{
					PrevBlockHash: dummyHash,
					MerkleRoot:    dummyHash,
					Timestamp:     12345,
				},
				Transactions: []*pb.Transaction{},
			},
		}

		_, err := server.Block(context.Background(), blockMsg)
		if err != nil {
			t.Fatalf("Block failed: %v", err)
		}

		select {
		case msg := <-mockObs.BlockCh:
			if msg.Block == nil {
				t.Error("Expected BlockMsg to contain a Block")
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for Block notification")
		}
	})

	t.Run("MerkleBlock", func(t *testing.T) {
		server, mockObs := setupTestServer()
		merkleBlockMsg := &pb.MerkleBlockMsg{
			MerkleBlock: &pb.MerkleBlock{
				Header: &pb.BlockHeader{
					PrevBlockHash: dummyHash,
					MerkleRoot:    dummyHash,
				},
				Proofs: []*pb.MerkleProof{},
			},
		}

		_, err := server.MerkleBlock(context.Background(), merkleBlockMsg)
		if err != nil {
			t.Fatalf("MerkleBlock failed: %v", err)
		}

		select {
		case msg := <-mockObs.MerkleBlockCh:
			if msg.MerkleBlock == nil {
				t.Error("Expected MerkleBlockMsg to contain a MerkleBlock")
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for MerkleBlock notification")
		}
	})

	t.Run("Tx", func(t *testing.T) {
		server, mockObs := setupTestServer()
		txMsg := &pb.TxMsg{
			Transaction: &pb.Transaction{
				Inputs:   []*pb.TxInput{},
				Outputs:  []*pb.TxOutput{},
				LockTime: 0,
			},
		}

		_, err := server.Tx(context.Background(), txMsg)
		if err != nil {
			t.Fatalf("Tx failed: %v", err)
		}

		select {
		case msg := <-mockObs.TxCh:
			if msg.Transaction == nil {
				t.Error("Expected TxMsg to contain a Transaction")
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for Tx notification")
		}
	})

	t.Run("GetHeaders", func(t *testing.T) {
		server, mockObs := setupTestServer()
		locatorMsg := &pb.BlockLocator{
			BlockLocatorHashes: [][]byte{dummyHash},
			HashStop:           dummyHash,
		}

		_, err := server.GetHeaders(context.Background(), locatorMsg)
		if err != nil {
			t.Fatalf("GetHeaders failed: %v", err)
		}

		select {
		case locator := <-mockObs.GetHeadersCh:
			if len(locator.BlockLocatorHashes) != 1 {
				t.Errorf("Expected 1 hash in locator, got %d", len(locator.BlockLocatorHashes))
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for GetHeaders notification")
		}
	})

	t.Run("Headers", func(t *testing.T) {
		server, mockObs := setupTestServer()
		headersMsg := &pb.BlockHeaders{
			Headers: []*pb.BlockHeader{
				{
					PrevBlockHash: dummyHash,
					MerkleRoot:    dummyHash,
				},
			},
		}

		_, err := server.Headers(context.Background(), headersMsg)
		if err != nil {
			t.Fatalf("Headers failed: %v", err)
		}

		select {
		case headers := <-mockObs.HeadersCh:
			if len(headers) != 1 {
				t.Errorf("Expected 1 header, got %d", len(headers))
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for Headers notification")
		}
	})

	t.Run("SetFilter", func(t *testing.T) {
		server, mockObs := setupTestServer()
		filterMsg := &pb.SetFilterRequest{
			PublicKeyHashes: [][]byte{dummyHash},
		}

		_, err := server.SetFilter(context.Background(), filterMsg)
		if err != nil {
			t.Fatalf("SetFilter failed: %v", err)
		}

		select {
		case req := <-mockObs.SetFilterCh:
			if len(req.PublicKeyHashes) != 1 {
				t.Errorf("Expected 1 public key hash, got %d", len(req.PublicKeyHashes))
			}
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for SetFilter notification")
		}
	})

	t.Run("Mempool", func(t *testing.T) {
		server, mockObs := setupTestServer()

		_, err := server.Mempool(context.Background(), nil)
		if err != nil {
			t.Fatalf("Mempool failed: %v", err)
		}

		select {
		case <-mockObs.MempoolCh:
			// Success
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for Mempool notification")
		}
	})
}
