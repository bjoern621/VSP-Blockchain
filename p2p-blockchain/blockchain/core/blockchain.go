package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/miner/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	mapset "github.com/deckarep/golang-set/v2"
)

const invalidBlockMessageFormat = "Block Message received from %v is invalid: %v"

type Blockchain struct {
	mempool                *Mempool
	blockchainMsgSender    api.BlockchainAPI
	fullInventoryMsgSender api.FullInventoryInformationMsgSenderAPI

	transactionValidator validation.ValidationAPI
	blockValidator       validation.BlockValidationAPI

	blockStore          blockchain.BlockStoreAPI
	chainReorganization ChainReorganizationAPI

	observers mapset.Set[observer.BlockchainObserverAPI]
}

func NewBlockchain(
	blockchainMsgSender api.BlockchainAPI,
	fullInventoryMsgSender api.FullInventoryInformationMsgSenderAPI,
	transactionValidator validation.ValidationAPI,
	blockValidator validation.BlockValidationAPI,
	blockStore blockchain.BlockStoreAPI,
	utxoService utxo.UTXOService,
) *Blockchain {
	mempool := NewMempool(transactionValidator, blockStore)
	return &Blockchain{
		mempool:                mempool,
		blockchainMsgSender:    blockchainMsgSender,
		fullInventoryMsgSender: fullInventoryMsgSender,

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
