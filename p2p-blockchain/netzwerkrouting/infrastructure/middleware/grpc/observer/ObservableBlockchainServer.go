package observer

import "s3b/vsp-blockchain/p2p-blockchain/blockchain"

type ObservableBlockchainServer interface {
	Attach(o *BlockchainObserver) (bool, error)
	Detach(o *BlockchainObserver) (bool, error)

	NotifyInv(invMsg blockchain.InvMsg)
}
