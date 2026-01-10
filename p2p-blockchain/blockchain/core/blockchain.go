package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/miner/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/assert"
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
	mempool := NewMempool(transactionValidator)
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

func (b *Blockchain) Inv(inventory []*inv.InvVector, peerID common.PeerId) {
	logger.Infof("Inv Message received: %v from %v", inventory, peerID)

	unknownData := make([]*inv.InvVector, 0)

	for _, v := range inventory {
		switch v.InvType {
		case inv.InvTypeMsgBlock:
			if _, err := b.blockStore.GetBlockByHash(v.Hash); err != nil {
				unknownData = append(unknownData, v)
			}
		case inv.InvTypeMsgTx:
			if !b.mempool.IsKnownTransactionHash(v.Hash) || !b.IsTransactionKnown(v.Hash) {
				unknownData = append(unknownData, v)
			}
		case inv.InvTypeMsgFilteredBlock:
			panic("not implemented")
		}
	}

	b.requestData(unknownData, peerID)
}

func (b *Blockchain) GetData(inventory []*inv.InvVector, peerID common.PeerId) {
	logger.Infof("GetData Message received: %v from %v", inventory, peerID)
}

func (b *Blockchain) Block(receivedBlock block.Block, peerID common.PeerId) {
	// 1. Basic validation
	if ok, err := b.blockValidator.SanityCheck(receivedBlock); !ok {
		logger.Warnf(invalidBlockMessageFormat, peerID, err)
		return
	}

	if ok, err := b.blockValidator.ValidateHeader(receivedBlock); !ok {
		logger.Warnf(invalidBlockMessageFormat, peerID, err)
		return
	}

	b.NotifyStopMining()

	// 2. Add block to store
	addedBlocks := b.blockStore.AddBlock(receivedBlock)

	// 3. Handle orphans
	if isOrphan, err := b.blockStore.IsOrphanBlock(receivedBlock); isOrphan {
		logger.Infof("Block is Orphan: %v", err)
		assert.Assert(peerID != "", "Mined blocks should never be orphans")
		b.requestMissingBlockHeaders(receivedBlock, peerID)
		return
	}

	// 4. Full validation BEFORE applying to UTXO set
	if ok, _ := b.blockValidator.FullValidation(receivedBlock); !ok {
		// An invalid block will be "removed" in the block store in form of not beeing available for retreval
		return
	}

	// 5. Check if chain reorganization is needed
	tip := b.blockStore.GetMainChainTip()
	tipHash := tip.Hash()
	reorganized, err := b.chainReorganization.CheckAndReorganize(tipHash)
	if err != nil {
		logger.Errorf("Chain reorganization failed: %v", err)
		return
	}

	if reorganized {
		logger.Infof("Chain reorganization performed")
	}

	// 6. Broadcast new blocks
	b.broadcastAddedBlocks(addedBlocks, peerID)

	b.NotifyStartMining()

	logger.Infof("Block Message received: %v from %v", receivedBlock, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId) {
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlock, peerID)
}

// Tx processes a transaction message
// If the transaction is valid and not yet known, it is added to the mempool and broadcasted to other peers
func (b *Blockchain) Tx(tx transaction.Transaction, peerID common.PeerId) {

	isValid, err := b.transactionValidator.ValidateTransaction(&tx)
	if !isValid {
		logger.Errorf("Tx Message received from %v is invalid: %v", peerID, err)
		return
	}
	if b.mempool.IsKnownTransactionId(tx.TransactionId()) || b.IsTransactionKnownById(tx.TransactionId()) {
		logger.Infof("Tx Message already known: %v from %v", tx, peerID)
		return
	}

	isNew := b.mempool.AddTransaction(tx)
	if isNew {
		invVectors := make([]*inv.InvVector, 0)
		invVector := inv.InvVector{
			Hash:    tx.Hash(),
			InvType: inv.InvTypeMsgTx,
		}
		invVectors = append(invVectors, &invVector)
		b.blockchainMsgSender.BroadcastInvExclusionary(invVectors, peerID)
	}

	logger.Infof("Tx Message received: %v from %v", tx, peerID)
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

func (b *Blockchain) broadcastAddedBlocks(addedBlocks []common.Hash, peerID common.PeerId) {
	b.blockchainMsgSender.BroadcastAddedBlocks(addedBlocks, peerID)
}

func (b *Blockchain) requestMissingBlockHeaders(receivedBlock block.Block, peerId common.PeerId) {
	parentHash := receivedBlock.Header.PreviousBlockHash

	currentHeight := b.blockStore.GetCurrentHeight()
	locatorHashes := b.buildBlockLocator(currentHeight)

	// Prepend the orphan parent hash at the beginning (most recent hash)
	locatorHashes = append([]common.Hash{parentHash}, locatorHashes...)

	locator := block.BlockLocator{
		BlockLocatorHashes: locatorHashes,
		StopHash:           common.Hash{}, // Empty stop hash means don't stop until we find common ancestor
	}

	b.blockchainMsgSender.RequestMissingBlockHeaders(locator, peerId)
}

// buildBlockLocator creates a block locator using Fibonacci series to sample the chain.
// Returns hashes starting from newer blocks (closer to tip) to older blocks (closer to genesis).
func (b *Blockchain) buildBlockLocator(tipHeight uint64) []common.Hash {
	locatorHashes := make([]common.Hash, 0)

	fib1, fib2 := uint64(1), uint64(2)
	offset := uint64(0)

	for offset <= tipHeight {
		height := tipHeight - offset

		blocksAtHeight := b.blockStore.GetBlocksByHeight(height)

		for _, blk := range blocksAtHeight {
			if b.blockStore.IsPartOfMainChain(blk) {
				locatorHashes = append(locatorHashes, blk.Hash())
				break
			}
		}

		offset += fib1
		fib1, fib2 = fib2, fib1+fib2

		if len(locatorHashes) > 1000 {
			break
		}
	}

	return locatorHashes
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
