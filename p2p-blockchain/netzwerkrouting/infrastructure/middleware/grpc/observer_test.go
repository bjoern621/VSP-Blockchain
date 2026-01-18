package grpc

import (
	"reflect"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/observer"

	mapset "github.com/deckarep/golang-set/v2"
)

// MockObserver struct used to verify that the Server correctly notifies observers.
type MockObserver struct {
	InvCh         chan []*inv.InvVector
	GetDataCh     chan []*inv.InvVector
	BlockCh       chan block.Block
	MerkleBlockCh chan block.MerkleBlock
	TxCh          chan transaction.Transaction
	GetHeadersCh  chan block.BlockLocator
	HeadersCh     chan []*block.BlockHeader
	SetFilterCh   chan block.SetFilterRequest
	MempoolCh     chan struct{}
}

// NewMockObserver creates a MockObserver with buffered channels to prevent blocking during tests.
func NewMockObserver() *MockObserver {
	return &MockObserver{
		InvCh:         make(chan []*inv.InvVector, 10),
		GetDataCh:     make(chan []*inv.InvVector, 10),
		BlockCh:       make(chan block.Block, 10),
		MerkleBlockCh: make(chan block.MerkleBlock, 10),
		TxCh:          make(chan transaction.Transaction, 10),
		GetHeadersCh:  make(chan block.BlockLocator, 10),
		HeadersCh:     make(chan []*block.BlockHeader, 10),
		SetFilterCh:   make(chan block.SetFilterRequest, 10),
		MempoolCh:     make(chan struct{}, 10),
	}
}

// Implement the BlockchainObserverAPI interface

func (m *MockObserver) Inv(inventory []*inv.InvVector, _ common.PeerId) {
	m.InvCh <- inventory
}

func (m *MockObserver) GetData(inventory []*inv.InvVector, _ common.PeerId) {
	m.GetDataCh <- inventory
}

func (m *MockObserver) Block(block block.Block, _ common.PeerId) {
	m.BlockCh <- block
}

func (m *MockObserver) MerkleBlock(merkleBlock block.MerkleBlock, _ common.PeerId) {
	m.MerkleBlockCh <- merkleBlock
}

func (m *MockObserver) Tx(tx transaction.Transaction, _ common.PeerId) {
	m.TxCh <- tx
}

func (m *MockObserver) GetHeaders(locator block.BlockLocator, _ common.PeerId) {
	m.GetHeadersCh <- locator
}

func (m *MockObserver) Headers(headers []*block.BlockHeader, _ common.PeerId) {
	m.HeadersCh <- headers
}

func (m *MockObserver) SetFilter(setFilterRequest block.SetFilterRequest, _ common.PeerId) {
	m.SetFilterCh <- setFilterRequest
}

func (m *MockObserver) Mempool(_ common.PeerId) {
	m.MempoolCh <- struct{}{}
}

func mustHash(b byte) common.Hash {
	var h common.Hash
	for i := 0; i < len(h); i++ {
		h[i] = b
	}
	return h
}

func mustPublicKeyHash(b byte) block.PublicKeyHash {
	var h block.PublicKeyHash
	for i := 0; i < len(h); i++ {
		h[i] = b
	}
	return h
}

// Tests

func TestObserverBlockchainServer_Notify(t *testing.T) {
	// 1. Create the Mock Observer
	mockObs := NewMockObserver()

	// 2. Instantiate the Server manually.
	// We avoid NewServer() or gRPC logic because we only want to test the observer pattern logic.
	// We manually initialize the 'observers' map since we are in the same package (grpc).
	server := &Server{
		observers: mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	// 3. Attach the mock observer
	// We can use server.Attach(mockObs) if available, or inject directly:
	server.observers.Add(mockObs)

	// Shared test data
	testPeerID := common.PeerId("test-peer-123")
	timeout := time.Millisecond * 100

	t.Run("NotifyInv", func(t *testing.T) {
		invVector := inv.InvVector{
			InvType: inv.InvTypeMsgBlock,
			Hash:    mustHash(0xAB),
		}
		inventory := make([]*inv.InvVector, 1)
		inventory[0] = &invVector

		server.NotifyInv(inventory, testPeerID)

		select {
		case received := <-mockObs.InvCh:
			if !reflect.DeepEqual(received, inventory) {
				t.Errorf("Expected message %v, got %v", inventory, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyInv")
		}
	})

	t.Run("NotifyGetData", func(t *testing.T) {
		invVector := inv.InvVector{
			InvType: inv.InvTypeMsgTx,
			Hash:    mustHash(0xCD),
		}
		inventory := make([]*inv.InvVector, 1)
		inventory[0] = &invVector

		server.NotifyGetData(inventory, testPeerID)

		select {
		case received := <-mockObs.GetDataCh:
			if !reflect.DeepEqual(received, inventory) {
				t.Errorf("Expected message %v, got %v", inventory, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyGetData")
		}
	})

	t.Run("NotifyBlock", func(t *testing.T) {
		b := block.Block{
			Header: block.BlockHeader{
				PreviousBlockHash: mustHash(0x01),
				MerkleRoot:        mustHash(0x02),
				Timestamp:         123,
				DifficultyTarget:  0,
				Nonce:             123,
			},
			Transactions: []transaction.Transaction{},
		}

		server.NotifyBlock(b, testPeerID)

		select {
		case received := <-mockObs.BlockCh:
			if !reflect.DeepEqual(received, b) {
				t.Errorf("Expected message %v, got %v", b, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyBlock")
		}
	})

	t.Run("NotifyMerkleBlock", func(t *testing.T) {
		mkBlock := block.MerkleBlock{
			BlockHeader: block.BlockHeader{
				PreviousBlockHash: mustHash(0x03),
				MerkleRoot:        mustHash(0x04),
				Timestamp:         456,
				DifficultyTarget:  0,
				Nonce:             456,
			},
			Proofs: []block.MerkleProof{{
				Transaction: transaction.Transaction{
					Inputs:  []transaction.Input{},
					Outputs: []transaction.Output{},
				},
				Siblings: []common.Hash{mustHash(0x10), mustHash(0x11)},
				Index:    1,
			}},
		}

		server.NotifyMerkleBlock(mkBlock, testPeerID)

		select {
		case received := <-mockObs.MerkleBlockCh:
			if !reflect.DeepEqual(received, mkBlock) {
				t.Errorf("Expected message %v, got %v", mkBlock, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyMerkleBlock")
		}
	})

	t.Run("NotifyTx", func(t *testing.T) {
		tx := transaction.Transaction{
			Inputs: []transaction.Input{{
				PrevTxID:    transaction.TransactionID(mustHash(0x22)),
				OutputIndex: 0,
				Signature:   []byte{0x30, 0x01},
				PubKey:      transaction.PubKey{},
			}},
			Outputs: []transaction.Output{{
				Value:      500,
				PubKeyHash: transaction.PubKeyHash{},
			}},
		}

		server.NotifyTx(tx, testPeerID)

		select {
		case received := <-mockObs.TxCh:
			if !reflect.DeepEqual(received, tx) {
				t.Errorf("Expected message %v, got %v", tx, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyTx")
		}
	})

	t.Run("NotifyGetHeaders", func(t *testing.T) {
		locator := block.BlockLocator{
			BlockLocatorHashes: []common.Hash{mustHash(0x33), mustHash(0x34)},
			StopHash:           mustHash(0x35),
		}

		server.NotifyGetHeaders(locator, testPeerID)

		select {
		case received := <-mockObs.GetHeadersCh:
			if !reflect.DeepEqual(received, locator) {
				t.Errorf("Expected message %v, got %v", locator, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyGetHeaders")
		}
	})

	t.Run("NotifyHeaders", func(t *testing.T) {
		h1 := &block.BlockHeader{
			PreviousBlockHash: mustHash(0x40),
			MerkleRoot:        mustHash(0x41),
			Timestamp:         789,
			DifficultyTarget:  0,
			Nonce:             789,
		}
		h2 := &block.BlockHeader{
			PreviousBlockHash: mustHash(0x42),
			MerkleRoot:        mustHash(0x43),
			Timestamp:         101,
			DifficultyTarget:  0,
			Nonce:             101,
		}
		headers := []*block.BlockHeader{
			h1,
			h2,
		}

		server.NotifyHeaders(headers, testPeerID)

		select {
		case received := <-mockObs.HeadersCh:
			if !reflect.DeepEqual(received, headers) {
				t.Errorf("Expected message %v, got %v", headers, received)
			}
		case <-time.After(timeout):
			t.Fatal("Timeout waiting for NotifyHeaders")
		}
	})

	t.Run("NotifySetFilterRequest", func(t *testing.T) {
		msg := block.SetFilterRequest{
			PublicKeyHashes: []block.PublicKeyHash{
				mustPublicKeyHash(0x55),
				mustPublicKeyHash(0x56),
			},
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
