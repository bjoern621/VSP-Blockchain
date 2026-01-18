package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/miner/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
	mapset "github.com/deckarep/golang-set/v2"
)

type peerRetriever interface {
	GetPeer(id common.PeerId) (*common.Peer, bool)
}

type Blockchain struct {
	mempool                *Mempool
	blockchainMsgSender    api.BlockchainAPI
	fullInventoryMsgSender api.FullInventoryInformationMsgSenderAPI
	errorMsgSender         api.ErrorMsgSenderAPI

	transactionValidator validation.TransactionValidatorAPI
	blockValidator       validation.BlockValidationAPI

	blockStore          blockchain.BlockStoreAPI
	chainReorganization ChainReorganizationAPI

	observers mapset.Set[observer.BlockchainObserverAPI]

	peerRetriever peerRetriever
}

func NewBlockchain(
	blockchainMsgSender api.BlockchainAPI,
	fullInventoryMsgSender api.FullInventoryInformationMsgSenderAPI,
	errorMsgSender api.ErrorMsgSenderAPI,
	blockValidator validation.BlockValidationAPI,
	blockStore blockchain.BlockStoreAPI,
	peerRetriever peerRetriever,
	transactionValidator validation.TransactionValidatorAPI,
	utxoService utxo.UtxoStoreAPI,
) *Blockchain {
	mempool := NewMempool(transactionValidator, blockStore)
	genesis := blockchain.GenesisBlock()
	genesisHash := genesis.Hash()
	return &Blockchain{
		mempool:                mempool,
		blockchainMsgSender:    blockchainMsgSender,
		fullInventoryMsgSender: fullInventoryMsgSender,
		errorMsgSender:         errorMsgSender,

		transactionValidator: transactionValidator,
		blockValidator:       blockValidator,

		blockStore:          blockStore,
		chainReorganization: NewChainReorganization(blockStore, utxoService, mempool, genesisHash),

		observers: mapset.NewSet[observer.BlockchainObserverAPI](),

		peerRetriever: peerRetriever,
	}
}

func (b *Blockchain) AddSelfMinedBlock(selfMinedBlock block.Block) {
	b.Block(selfMinedBlock, "")
}

func (b *Blockchain) SetFilter(_ block.SetFilterRequest, _ common.PeerId) {
	panic("No longer supported and will be removed later")
}

func (b *Blockchain) MerkleBlock(_ block.MerkleBlock, _ common.PeerId) {
	panic("No longer supported and will be removed later")
}

func (b *Blockchain) Attach(o observer.BlockchainObserverAPI) {
	b.observers.Add(o)
}

func (b *Blockchain) Detach(o observer.BlockchainObserverAPI) {
	b.observers.Remove(o)
}

func (b *Blockchain) NotifyStartMining() {
	transactions := b.mempool.GetTransactionsForMining()
	for o := range b.observers.Iter() {
		o.StartMining(transactions)
	}
}

func (b *Blockchain) NotifyStopMining() {
	for o := range b.observers.Iter() {
		o.StopMining()
	}
}

// CheckPeerIsConnected checks if the peer with the given ID exists and is in the connected state.
// Should be used at the beginning of message handlers to validate the peer.
func (b *Blockchain) CheckPeerIsConnected(peerID common.PeerId) bool {
	if peerID == "" {
		return true // Local peer ("local miner")
	}

	peer, exists := b.peerRetriever.GetPeer(peerID)
	if !exists {
		logger.Warnf("[blockchain] Peer %v not found", peerID)
		return false
	}

	if peer.State != common.StateConnected {
		logger.Warnf("[blockchain] Peer %s is not connected (state: %v)", peerID, peer.State)
		b.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectNotConnected, "unknown", []byte("sending peer is not in state connected"))
		return false
	}

	return true
}
