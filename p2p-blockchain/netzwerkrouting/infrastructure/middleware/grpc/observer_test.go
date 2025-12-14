package grpc

import (
	"reflect"
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// MockObserver struct used to verify that the Server correctly notifies observers.
type MockObserver struct {
	InvCh         chan dto.InvMsgDTO
	GetDataCh     chan dto.GetDataMsgDTO
	BlockCh       chan dto.BlockMsgDTO
	MerkleBlockCh chan dto.MerkleBlockMsgDTO
	TxCh          chan dto.TxMsgDTO
	GetHeadersCh  chan dto.BlockLocatorDTO
	HeadersCh     chan dto.BlockHeadersDTO
	SetFilterCh   chan dto.SetFilterRequestDTO
	MempoolCh     chan struct{}
}

// NewMockObserver creates a MockObserver with buffered channels to prevent blocking during tests.
func NewMockObserver() *MockObserver {
	return &MockObserver{
		InvCh:         make(chan dto.InvMsgDTO, 10),
		GetDataCh:     make(chan dto.GetDataMsgDTO, 10),
		BlockCh:       make(chan dto.BlockMsgDTO, 10),
		MerkleBlockCh: make(chan dto.MerkleBlockMsgDTO, 10),
		TxCh:          make(chan dto.TxMsgDTO, 10),
		GetHeadersCh:  make(chan dto.BlockLocatorDTO, 10),
		HeadersCh:     make(chan dto.BlockHeadersDTO, 10),
		SetFilterCh:   make(chan dto.SetFilterRequestDTO, 10),
		MempoolCh:     make(chan struct{}, 10),
	}
}

// Implement the BlockchainObserverAPI interface

func (m *MockObserver) Inv(invMsg dto.InvMsgDTO, _ peer.PeerID) {
	m.InvCh <- invMsg
}

func (m *MockObserver) GetData(getDataMsg dto.GetDataMsgDTO, _ peer.PeerID) {
	m.GetDataCh <- getDataMsg
}

func (m *MockObserver) Block(blockMsg dto.BlockMsgDTO, _ peer.PeerID) {
	m.BlockCh <- blockMsg
}

func (m *MockObserver) MerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, _ peer.PeerID) {
	m.MerkleBlockCh <- merkleBlockMsg
}

func (m *MockObserver) Tx(txMsg dto.TxMsgDTO, _ peer.PeerID) {
	m.TxCh <- txMsg
}

func (m *MockObserver) GetHeaders(locator dto.BlockLocatorDTO, _ peer.PeerID) {
	m.GetHeadersCh <- locator
}

func (m *MockObserver) Headers(headers dto.BlockHeadersDTO, _ peer.PeerID) {
	m.HeadersCh <- headers
}

func (m *MockObserver) SetFilter(setFilterRequest dto.SetFilterRequestDTO, _ peer.PeerID) {
	m.SetFilterCh <- setFilterRequest
}

func (m *MockObserver) Mempool(_ peer.PeerID) {
	m.MempoolCh <- struct{}{}
}

func mustHash(b byte) dto.Hash {
	var h dto.Hash
	for i := 0; i < len(h); i++ {
		h[i] = b
	}
	return h
}

func mustPublicKeyHash(b byte) dto.PublicKeyHash {
	var h dto.PublicKeyHash
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
		observers: make(map[observer.BlockchainObserverAPI]struct{}),
	}

	// 3. Attach the mock observer
	// We can use server.Attach(mockObs) if available, or inject directly:
	server.observers[mockObs] = struct{}{}

	// Shared test data
	testPeerID := peer.PeerID("test-peer-123")
	timeout := time.Millisecond * 100

	t.Run("NotifyInv", func(t *testing.T) {
		msg := dto.InvMsgDTO{
			Inventory: []dto.InvVectorDTO{{
				Type: dto.InvTypeDTO_MSG_BLOCK,
				Hash: mustHash(0xAB),
			}},
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
		msg := dto.GetDataMsgDTO{
			Inventory: []dto.InvVectorDTO{{
				Type: dto.InvTypeDTO_MSG_TX,
				Hash: mustHash(0xCD),
			}},
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
		msg := dto.BlockMsgDTO{
			Block: dto.BlockDTO{
				Header: dto.BlockHeaderDTO{
					PrevBlockHash:    mustHash(0x01),
					MerkleRoot:       mustHash(0x02),
					Timestamp:        123,
					DifficultyTarget: 0x1d00ffff,
					Nonce:            123,
				},
				Transactions: []dto.TransactionDTO{},
			},
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
		msg := dto.MerkleBlockMsgDTO{
			MerkleBlock: dto.MerkleBlockDTO{
				Header: dto.BlockHeaderDTO{
					PrevBlockHash:    mustHash(0x03),
					MerkleRoot:       mustHash(0x04),
					Timestamp:        456,
					DifficultyTarget: 0x1d00ffff,
					Nonce:            456,
				},
				Proofs: []dto.MerkleProofDTO{{
					Transaction: dto.TransactionDTO{
						Inputs:   []dto.TxInputDTO{},
						Outputs:  []dto.TxOutputDTO{},
						LockTime: 0,
					},
					Siblings: []dto.Hash{mustHash(0x10), mustHash(0x11)},
					Index:    1,
				}},
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
		msg := dto.TxMsgDTO{
			Transaction: dto.TransactionDTO{
				Inputs: []dto.TxInputDTO{{
					PrevTxHash:      mustHash(0x22),
					OutputIndex:     0,
					SignatureScript: []byte{0x30, 0x01},
					Sequence:        0xffffffff,
				}},
				Outputs: []dto.TxOutputDTO{{
					Value:           500,
					PublicKeyScript: []byte{0x76, 0xa9},
				}},
				LockTime: 500,
			},
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
		msg := dto.BlockLocatorDTO{
			BlockLocatorHashes: []dto.Hash{mustHash(0x33), mustHash(0x34)},
			HashStop:           mustHash(0x35),
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
		msg := dto.BlockHeadersDTO{
			Headers: []dto.BlockHeaderDTO{
				{
					PrevBlockHash:    mustHash(0x40),
					MerkleRoot:       mustHash(0x41),
					Timestamp:        789,
					DifficultyTarget: 0x1d00ffff,
					Nonce:            789,
				},
				{
					PrevBlockHash:    mustHash(0x42),
					MerkleRoot:       mustHash(0x43),
					Timestamp:        101,
					DifficultyTarget: 0x1d00ffff,
					Nonce:            101,
				},
			},
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
		msg := dto.SetFilterRequestDTO{
			PublicKeyHashes: []dto.PublicKeyHash{
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
