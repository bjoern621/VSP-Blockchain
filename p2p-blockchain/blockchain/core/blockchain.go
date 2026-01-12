package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/miner/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
	mapset "github.com/deckarep/golang-set/v2"
)

const invalidBlockMessageFormat = "Block Message received from %v is invalid: %v"

type Blockchain struct {
	mempool             *Mempool
	blockchainMsgSender api.BlockchainAPI

	transactionValidator validation.ValidationAPI
	blockValidator       validation.BlockValidationAPI

	blockStore          blockchain.BlockStoreAPI
	chainReorganization ChainReorganizationAPI

	observers mapset.Set[observer.BlockchainObserverAPI]
}

func NewBlockchain(
	blockchainMsgSender api.BlockchainAPI,
	transactionValidator validation.ValidationAPI,
	blockValidator validation.BlockValidationAPI,
	blockStore blockchain.BlockStoreAPI,
	utxoService utxo.UTXOService,
) *Blockchain {
	mempool := NewMempool(transactionValidator, blockStore)
	return &Blockchain{
		mempool:             mempool,
		blockchainMsgSender: blockchainMsgSender,

		transactionValidator: transactionValidator,
		blockValidator:       blockValidator,

		blockStore:          blockStore,
		chainReorganization: NewChainReorganization(blockStore, utxoService, mempool),

		observers: mapset.NewSet[observer.BlockchainObserverAPI](),
	}
}

func (b *Blockchain) AddSelfMinedBlock(selfMinedBlock block.Block) {
	b.Block(selfMinedBlock, "")
}

func (b *Blockchain) GetData(inventory []*inv.InvVector, peerID common.PeerId) {
	logger.Infof("GetData Message received: %v from %v", inventory, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId) {
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlock, peerID)
}

func (b *Blockchain) GetHeaders(locator block.BlockLocator, peerID common.PeerId) {
	logger.Infof("GetHeaders Message received: %v from %v", locator, peerID)
}

func (b *Blockchain) Headers(blockHeaders []*block.BlockHeader, peerID common.PeerId) {
	logger.Infof("Headers Message received: %v from %v", blockHeaders, peerID)
}

func (b *Blockchain) SetFilter(setFilterRequest block.SetFilterRequest, peerID common.PeerId) {
	logger.Infof("setFilerRequest Message received: %v from %v", setFilterRequest, peerID)
}
func (b *Blockchain) Mempool(peerID common.PeerId) {
	logger.Infof("Mempool Message received from %v", peerID)
}

func (b *Blockchain) requestData(missingData []*inv.InvVector, id common.PeerId) {
	b.blockchainMsgSender.SendGetData(missingData, id)
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
