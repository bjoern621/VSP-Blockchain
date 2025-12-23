package utxopool

import (
	"encoding/binary"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// OutpointKeySize is the size of a serialized outpoint key (32 bytes TxID + 4 bytes OutputIndex)
const OutpointKeySize = common.HashSize + 4

// Outpoint uniquely identifies a transaction output
type Outpoint struct {
	TxID        transaction.TransactionID
	OutputIndex uint32
}

// NewOutpoint creates a new Outpoint from a transaction ID and output index
func NewOutpoint(txID transaction.TransactionID, outputIndex uint32) Outpoint {
	return Outpoint{
		TxID:        txID,
		OutputIndex: outputIndex,
	}
}

// Key returns a byte slice key for storage lookup
// Format: TxID (32 bytes) + OutputIndex (4 bytes, big-endian)
func (o Outpoint) Key() []byte {
	key := make([]byte, OutpointKeySize)
	copy(key[:common.HashSize], o.TxID[:])
	binary.BigEndian.PutUint32(key[common.HashSize:], o.OutputIndex)
	return key
}

// OutpointFromKey reconstructs an Outpoint from a byte key
func OutpointFromKey(key []byte) Outpoint {
	var outpoint Outpoint
	copy(outpoint.TxID[:], key[:common.HashSize])
	outpoint.OutputIndex = binary.BigEndian.Uint32(key[common.HashSize:])
	return outpoint
}
