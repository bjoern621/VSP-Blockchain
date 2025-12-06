package observer

import "s3b/vsp-blockchain/p2p-blockchain/blockchain"

type BlockchainObserver interface {
	Inv(invMsg *blockchain.InvMsg)
	GetData(getDataMsg *blockchain.GetDataMsg)
	Block(blockMsg *blockchain.BlockMsg)
	MerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg)
	Tx(txMsg *blockchain.TxMsg)
	GetHeaders(locator *blockchain.BlockLocator)
	Headers(headers []*blockchain.BlockHeader)
	SetFilter(setFilterRequest *blockchain.SetFilterRequest)
}
