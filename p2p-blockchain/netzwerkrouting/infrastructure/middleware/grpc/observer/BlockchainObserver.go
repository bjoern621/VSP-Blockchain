package observer

import "s3b/vsp-blockchain/p2p-blockchain/blockchain"

type BlockchainObserver interface {
	Inv(invMsg blockchain.InvMsg)
	GetData(getDataMsg blockchain.GetDataMsg)
}
