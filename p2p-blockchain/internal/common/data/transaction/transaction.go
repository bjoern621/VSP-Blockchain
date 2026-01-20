package transaction

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Transaction struct {
	Inputs  []Input
	Outputs []Output
}

func NewTransaction(
	utxos []UTXO,
	toPubKeyHash PubKeyHash,
	amount uint64,
	fee uint64,
	privateKey PrivateKey,
) (*Transaction, error) {

	selected, total := selectUTXOs(utxos, amount+fee)
	if total < amount+fee {
		return &Transaction{}, ErrInsufficientFunds
	}
	change := total - amount - fee

	tx := fillInTransactionData(
		selected,
		amount,
		toPubKeyHash,
		privateKey,
		change,
	)

	if err := tx.Sign(privateKey, selected); err != nil {
		return nil, err
	}

	return tx, nil
}

func fillInTransactionData(selected []UTXO, amount uint64, toPubKeyHash PubKeyHash, privateKey PrivateKey, change uint64) *Transaction {
	tx := &Transaction{}

	tx.addUnsignedInputs(selected)

	tx.addOutput(amount, toPubKeyHash)
	if change > 0 {
		tx.addChange(change, privateKey)
	}
	return tx
}

func (tx *Transaction) addChange(change uint64, privateKey PrivateKey) {
	pubKey := pubFromPriv(privateKey)
	ownAddress := Hash160(pubKey)

	tx.Outputs = append(tx.Outputs, Output{
		Value:      change,
		PubKeyHash: ownAddress,
	})
}

func (tx *Transaction) addOutput(amount uint64, toPubKeyHash PubKeyHash) {
	tx.Outputs = append(tx.Outputs, Output{
		Value:      amount,
		PubKeyHash: toPubKeyHash,
	})
}

func (tx *Transaction) addUnsignedInputs(selected []UTXO) {
	for _, u := range selected {
		tx.addInput(u)
	}
}

func (tx *Transaction) addInput(u UTXO) {
	tx.Inputs = append(tx.Inputs, Input{
		PrevTxID:    u.TxID,
		OutputIndex: u.OutputIndex,
	})
}

func (tx *Transaction) Clone() *Transaction {
	clone := &Transaction{}

	// Deep copy inputs
	clone.Inputs = make([]Input, len(tx.Inputs))
	for i, in := range tx.Inputs {
		clone.Inputs[i] = in.Clone()
	}

	// Deep copy outputs
	clone.Outputs = make([]Output, len(tx.Outputs))
	for i, out := range tx.Outputs {
		clone.Outputs[i] = out.Clone()
	}

	return clone
}

// Hash computes the block.Hash similar to the bitcoin wtxid with transaction data including witness
func (tx *Transaction) Hash() common.Hash {
	txId := tx.TransactionId()
	var hash common.Hash
	copy(hash[:], txId[:])
	return hash
}

// TransactionId computes the TransactionID similar to the bitcoin wtxid with transaction data including witness
func (tx *Transaction) TransactionId() TransactionID {
	buf := serializeTransaction(tx)
	hash := doubleSHA256(buf.Bytes())

	var txID TransactionID
	copy(txID[:], hash)
	return txID
}

func serializeTransaction(tx *Transaction) *bytes.Buffer {
	buf := new(bytes.Buffer)
	serializeInputs(tx, buf)
	serializeOutputs(tx, buf)
	return buf
}

func serializeOutputs(tx *Transaction, buf *bytes.Buffer) {
	writeUint32(buf, uint32(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		writeUint64(buf, out.Value)
		buf.Write(out.PubKeyHash[:])
	}
}

func serializeInputs(tx *Transaction, buf *bytes.Buffer) {
	writeUint32(buf, uint32(len(tx.Inputs)))
	for _, in := range tx.Inputs {
		serializeInput(buf, in)
	}
}

func serializeInput(buf *bytes.Buffer, in Input) {
	buf.Write(in.PrevTxID[:])
	writeUint32(buf, in.OutputIndex)
	writeUint32(buf, uint32(len(in.Signature)))
	writeBytes(buf, in.Signature)
	buf.Write(in.PubKey[:])
}

func doubleSHA256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

// IsCoinbase returns true if the transaction is a coinbase transaction.
// A coinbase transaction has exactly one input with:
// - PrevTxID is all zeros (empty TransactionID)
// - OutputIndex is 0xFFFFFFFF
func (tx *Transaction) IsCoinbase() bool {
	if len(tx.Inputs) != 1 {
		return false
	}

	input := tx.Inputs[0]

	emptyTxID := TransactionID{}
	if input.PrevTxID != emptyTxID {
		return false
	}

	if input.OutputIndex != 0xFFFFFFFF {
		return false
	}

	return true
}

// NewCoinbaseTransaction creates a new coinbase transaction.
// A coinbase transaction is the first transaction in a block and rewards the miner.
//
// Parameters:
//   - receiverPubKeyHash: The 20-byte public key hash (address) of the miner who will receive the reward
//   - blockReward: The total amount to reward the miner (block subsidy + transaction fees)
//   - coinbaseData: Optional arbitrary data to include in the coinbase input (e.g., block height, nonce, or a message).
//     If empty, a default message will be used. Max 100 bytes recommended.
//
// The coinbase transaction has:
//   - One input with PrevTxID = 0, OutputIndex = 0xFFFFFFFF (coinbase specific)
//   - One or more outputs paying to the miner's address
//   - No signature verification needed (it's newly created coins)
func NewCoinbaseTransaction(receiverPubKeyHash PubKeyHash, blockReward uint64, height uint64) Transaction {
	// Build coinbase data (signature field can contain arbitrary data in coinbase)
	var signature [100]byte
	binary.LittleEndian.PutUint64(signature[:], height)

	rand.Read(signature[8:]) //nolint:errcheck

	return Transaction{
		Inputs: []Input{
			{
				PrevTxID:    TransactionID{},
				OutputIndex: 0xFFFFFFFF,
				Signature:   signature[:],
				PubKey:      PubKey{},
			},
		},
		Outputs: []Output{
			{
				Value:      blockReward,
				PubKeyHash: receiverPubKeyHash,
			},
		},
	}
}

func (tx *Transaction) String() string {
	hash := tx.Hash()
	return hex.EncodeToString(hash[:])
}
