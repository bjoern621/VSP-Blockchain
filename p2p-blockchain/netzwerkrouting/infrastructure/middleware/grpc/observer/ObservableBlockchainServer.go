package observer

import "s3b/vsp-blockchain/p2p-blockchain/blockchain"

type ObservableBlockchainServer interface {
	NotifyInv(invMsg blockchain.InvMsg)
}
