package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
)

type ObservableBlockchainServer interface {
	Attach(o *BlockchainObserver)
	Detach(o *BlockchainObserver)

	NotifyInv(invMsg blockchain.InvMsg)
	NotifyGetData(getDataMsg blockchain.GetDataMsg)
	NotifyBlock(blockMsg blockchain.BlockMsg)
	NotifyMerkleBlock(merkleBlockMsg blockchain.MerkleBlockMsg)
	NotifyTx(txMsg blockchain.TxMsg)
	NotifyGetHeaders(blockLocator blockchain.BlockLocator)
}
