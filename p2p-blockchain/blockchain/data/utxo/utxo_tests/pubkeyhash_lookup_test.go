package utxo_tests

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/infrastructure"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestChainState_GetUTXOsByPubKeyHash(t *testing.T) {
	// Create in-memory DAO for testing
	dao, err := infrastructure.NewUTXOEntryDAO(infrastructure.NewUTXOEntryDAOConfig("", true))
	if err != nil {
		t.Fatalf("Failed to create DAO: %v", err)
	}
	defer func(dao *infrastructure.UTXOEntryDAOImpl) {
		err = dao.Close()
		if err != nil {
			t.Fatalf("Failed to close dao: %v", err)
		}
	}(dao)

	chainstate, err := utxo.NewChainStateService(utxo.ChainStateConfig{CacheSize: 100}, dao)
	if err != nil {
		t.Fatalf("Failed to create chainstate: %v", err)
	}
	defer func(chainstate *utxo.ChainStateService) {
		err = chainstate.Close()
		if err != nil {
			t.Fatalf("Failed to close dao: %v", err)
		}
	}(chainstate)

	// Create test PubKeyHashes
	var pkh1, pkh2 transaction.PubKeyHash
	copy(pkh1[:], []byte("pubkeyhash1_________"))
	copy(pkh2[:], []byte("pubkeyhash2_________"))

	// Create test transaction IDs
	var txID1, txID2, txID3 transaction.TransactionID
	copy(txID1[:], []byte("tx1_____________________________"))
	copy(txID2[:], []byte("tx2_____________________________"))
	copy(txID3[:], []byte("tx3_____________________________"))

	// Add UTXOs for pkh1 (2 UTXOs)
	outpoint1 := utxopool.NewOutpoint(txID1, 0)
	entry1 := utxopool.NewUTXOEntry(transaction.Output{Value: 100, PubKeyHash: pkh1}, 1, false)
	if err = chainstate.Add(outpoint1, entry1); err != nil {
		t.Fatalf("Failed to add UTXO: %v", err)
	}

	outpoint2 := utxopool.NewOutpoint(txID2, 1)
	entry2 := utxopool.NewUTXOEntry(transaction.Output{Value: 200, PubKeyHash: pkh1}, 2, false)
	if err = chainstate.Add(outpoint2, entry2); err != nil {
		t.Fatalf("Failed to add UTXO: %v", err)
	}

	// Add UTXO for pkh2 (1 UTXO)
	outpoint3 := utxopool.NewOutpoint(txID3, 0)
	entry3 := utxopool.NewUTXOEntry(transaction.Output{Value: 300, PubKeyHash: pkh2}, 3, true)
	if err = chainstate.Add(outpoint3, entry3); err != nil {
		t.Fatalf("Failed to add UTXO: %v", err)
	}

	// Test: Get UTXOs for pkh1
	results, err2 := chainstate.GetUTXOsByPubKeyHash(pkh1)
	if err2 != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 UTXOs for pkh1, got %d", len(results))
	}

	// Verify values
	totalValue := uint64(0)
	for _, uwp := range results {
		totalValue += uwp.Output.Value
	}
	if totalValue != 300 { // 100 + 200
		t.Errorf("Expected total value 300, got %d", totalValue)
	}

	// Test: Get UTXOs for pkh2
	results2, err3 := chainstate.GetUTXOsByPubKeyHash(pkh2)
	if err3 != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("Expected 1 UTXO for pkh2, got %d", len(results2))
	}
	if results2[0].Output.Value != 300 {
		t.Errorf("Expected value 300, got %d", results2[0].Output.Value)
	}

	// Test: Get UTXOs for non-existent PubKeyHash
	var pkh3 transaction.PubKeyHash
	copy(pkh3[:], []byte("pubkeyhash3_________"))
	results3, err4 := chainstate.GetUTXOsByPubKeyHash(pkh3)
	if err4 != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results3) != 0 {
		t.Errorf("Expected 0 UTXOs for pkh3, got %d", len(results3))
	}

	// Test: Remove UTXO and verify index is updated
	if err = chainstate.Remove(outpoint1); err != nil {
		t.Fatalf("Failed to remove UTXO: %v", err)
	}
	results4, err5 := chainstate.GetUTXOsByPubKeyHash(pkh1)
	if err5 != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results4) != 1 {
		t.Errorf("Expected 1 UTXO for pkh1 after removal, got %d", len(results4))
	}
}

func TestMemPool_GetUTXOsByPubKeyHash(t *testing.T) {
	mempool := utxo.NewMemUTXOPool()
	defer func(mempool *utxo.MemPoolService) {
		err := mempool.Close()
		if err != nil {
			t.Fatalf("Failed to close mempool: %v", err)
		}
	}(mempool)

	// Create test PubKeyHashes
	var pkh1, pkh2 transaction.PubKeyHash
	copy(pkh1[:], []byte("pubkeyhash1_________"))
	copy(pkh2[:], []byte("pubkeyhash2_________"))

	// Create test transaction IDs
	var txID1, txID2 transaction.TransactionID
	copy(txID1[:], []byte("tx1_____________________________"))
	copy(txID2[:], []byte("tx2_____________________________"))

	// Add UTXOs
	outpoint1 := utxopool.NewOutpoint(txID1, 0)
	entry1 := utxopool.NewUTXOEntry(transaction.Output{Value: 100, PubKeyHash: pkh1}, 0, false)
	err := mempool.Add(outpoint1, entry1)
	if err != nil {
		return
	}

	outpoint2 := utxopool.NewOutpoint(txID2, 0)
	entry2 := utxopool.NewUTXOEntry(transaction.Output{Value: 200, PubKeyHash: pkh2}, 0, false)
	err = mempool.Add(outpoint2, entry2)
	if err != nil {
		return
	}

	// Test: Get UTXOs for pkh1
	results, err := mempool.GetUTXOsByPubKeyHash(pkh1)
	if err != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 UTXO for pkh1, got %d", len(results))
	}

	// Test: GetUTXOsWithOutpointByPubKeyHash
	resultsWithOutpoint, err := mempool.GetUTXOsWithOutpointByPubKeyHash(pkh1)
	if err != nil {
		t.Fatalf("GetUTXOsWithOutpointByPubKeyHash failed: %v", err)
	}
	if len(resultsWithOutpoint) != 1 {
		t.Errorf("Expected 1 UTXO for pkh1, got %d", len(resultsWithOutpoint))
	}
	if resultsWithOutpoint[0].Output.Value != 100 {
		t.Errorf("Expected value 100, got %d", resultsWithOutpoint[0].Output.Value)
	}
}

func TestFullNode_GetUTXOsByPubKeyHash(t *testing.T) {
	// Create in-memory DAO for testing
	dao, err := infrastructure.NewUTXOEntryDAO(infrastructure.NewUTXOEntryDAOConfig("", true))
	if err != nil {
		t.Fatalf("Failed to create DAO: %v", err)
	}

	chainstate, err4 := utxo.NewChainStateService(utxo.ChainStateConfig{CacheSize: 100}, dao)
	if err4 != nil {
		t.Fatalf("Failed to create chainstate: %v", err)
	}

	mempool := utxo.NewMemUTXOPool()
	fullNode := utxo.NewCombinedUTXOPool(mempool, chainstate)
	defer func(fullNode *utxo.FullNodeUTXOService) {
		err = fullNode.Close()
		if err != nil {
			t.Fatalf("Failed to close fullnode: %v", err)
		}
	}(fullNode)

	// Create test PubKeyHash
	var pkh1 transaction.PubKeyHash
	copy(pkh1[:], []byte("pubkeyhash1_________"))

	// Create test transaction IDs
	var txID1, txID2, txID3 transaction.TransactionID
	copy(txID1[:], []byte("tx1_____________________________"))
	copy(txID2[:], []byte("tx2_____________________________"))
	copy(txID3[:], []byte("tx3_____________________________"))

	// Add confirmed UTXO to chainstate
	outpoint1 := utxopool.NewOutpoint(txID1, 0)
	entry1 := utxopool.NewUTXOEntry(transaction.Output{Value: 100, PubKeyHash: pkh1}, 1, false)
	err = chainstate.Add(outpoint1, entry1)
	if err != nil {
		return
	}

	// Add unconfirmed UTXO to mempool
	outpoint2 := utxopool.NewOutpoint(txID2, 0)
	entry2 := utxopool.NewUTXOEntry(transaction.Output{Value: 200, PubKeyHash: pkh1}, 0, false)
	err = mempool.Add(outpoint2, entry2)
	if err != nil {
		return
	}

	// Test: Get all UTXOs (confirmed + unconfirmed)
	results, err2 := fullNode.GetUTXOsByPubKeyHash(pkh1)
	if err2 != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 UTXOs, got %d", len(results))
	}

	// Test: Mark chainstate UTXO as spent in mempool
	mempool.MarkSpent(outpoint1)

	// Should now only return the unconfirmed UTXO
	results2, err3 := fullNode.GetUTXOsByPubKeyHash(pkh1)
	if err3 != nil {
		t.Fatalf("GetUTXOsByPubKeyHash failed: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("Expected 1 UTXO after marking spent, got %d", len(results2))
	}
	if results2[0].Output.Value != 200 {
		t.Errorf("Expected unconfirmed UTXO with value 200, got %d", results2[0].Output.Value)
	}
}

func TestDAO_SecondaryIndex(t *testing.T) {
	// Create in-memory DAO for testing
	dao, err := infrastructure.NewUTXOEntryDAO(infrastructure.NewUTXOEntryDAOConfig("", true))
	if err != nil {
		t.Fatalf("Failed to create DAO: %v", err)
	}
	defer func(dao *infrastructure.UTXOEntryDAOImpl) {
		err = dao.Close()
		if err != nil {
			t.Fatalf("Failed to close dao: %v", err)
		}
	}(dao)

	// Create test PubKeyHash
	var pkh1 transaction.PubKeyHash
	copy(pkh1[:], []byte("pubkeyhash1_________"))

	// Create test transaction ID
	var txID1 transaction.TransactionID
	copy(txID1[:], []byte("tx1_____________________________"))

	outpoint := utxopool.NewOutpoint(txID1, 0)
	entry := utxopool.NewUTXOEntry(transaction.Output{Value: 100, PubKeyHash: pkh1}, 1, false)

	// Add UTXO
	if err = dao.Update(outpoint, entry); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify secondary index works
	outpoints, err := dao.FindByPubKeyHash(pkh1)
	if err != nil {
		t.Fatalf("FindByPubKeyHash failed: %v", err)
	}
	if len(outpoints) != 1 {
		t.Fatalf("Expected 1 outpoint, got %d", len(outpoints))
	}

	// Verify outpoint matches
	if outpoints[0].TxID != txID1 || outpoints[0].OutputIndex != 0 {
		t.Error("Returned outpoint does not match expected")
	}

	// Delete UTXO and verify index is cleaned up
	if err = dao.Delete(outpoint); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	outpoints2, err := dao.FindByPubKeyHash(pkh1)
	if err != nil {
		t.Fatalf("FindByPubKeyHash failed: %v", err)
	}
	if len(outpoints2) != 0 {
		t.Errorf("Expected 0 outpoints after delete, got %d", len(outpoints2))
	}
}
