package core

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	blockapi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

const (
	// graphvizOnlineBaseURL is the base URL for the GraphvizOnline visualization tool.
	graphvizOnlineBaseURL = "https://dreampuf.github.io/GraphvizOnline/?engine=dot#"

	// shortHashLength is the number of hex characters to use for shortened hash display.
	shortHashLength = 10

	// Node colors for different block types
	colorMainChain = "#90EE90" // Light green for main chain
	colorSideChain = "#ffffcc" // Light yellow for side chain
	colorOrphan    = "#ffcccc" // Light red for orphans
)

// VisualizationService generates blockchain visualizations.
type VisualizationService struct {
	blockStore blockapi.BlockStoreVisualizationAPI
}

// NewVisualizationService creates a new VisualizationService with the given block store.
func NewVisualizationService(blockStore blockapi.BlockStoreVisualizationAPI) *VisualizationService {
	return &VisualizationService{
		blockStore: blockStore,
	}
}

// GetVisualizationURL returns a URL to GraphvizOnline that displays the blockchain structure.
// If includeDetails is true, nodes will include height, accumulated work, and transaction tooltips.
func (v *VisualizationService) GetVisualizationURL(includeDetails bool) string {
	dotContent := v.generateDotContent(includeDetails)

	// URL-encode the DOT content and create the GraphvizOnline URL
	encodedDot := url.PathEscape(dotContent)
	return graphvizOnlineBaseURL + encodedDot
}

// generateDotContent generates the raw DOT format string for the blockchain.
func (v *VisualizationService) generateDotContent(includeDetails bool) string {
	var sb strings.Builder

	sb.WriteString("digraph Blockchain {\n")
	sb.WriteString("    rankdir=TB;\n")
	sb.WriteString("    node [shape=box, style=filled];\n")
	sb.WriteString("    edge [dir=back];\n\n")

	blocks := v.blockStore.GetAllBlocksWithMetadata()

	for _, bm := range blocks {
		v.generateDotForBlock(&sb, &bm, includeDetails)
	}

	sb.WriteString("}\n")

	return sb.String()
}

// generateDotForBlock generates DOT format output for a single block.
func (v *VisualizationService) generateDotForBlock(sb *strings.Builder, bm *block.BlockWithMetadata, includeDetails bool) {
	blockHash := bm.Block.Hash()
	hashStr := hex.EncodeToString(blockHash[:])
	shortHash := shortenHash(hashStr)

	fillColor := getNodeColor(bm.IsMainChain, bm.IsOrphan)

	var label string
	if includeDetails {
		label = fmt.Sprintf("%s\\nH:%d W:%d", shortHash, bm.Height, bm.AccumulatedWork)
	} else {
		label = shortHash
	}

	// Build node definition with optional tooltip
	if includeDetails {
		tooltip := formatBlockTooltip(&bm.Block)
		fmt.Fprintf(sb, "    \"%s\" [label=\"%s\", fillcolor=\"%s\", tooltip=\"%s\"];\n", hashStr, label, fillColor, tooltip)
	} else {
		fmt.Fprintf(sb, "    \"%s\" [label=\"%s\", fillcolor=\"%s\"];\n", hashStr, label, fillColor)
	}

	// Write edge from this node to parent
	if bm.ParentHash != nil {
		parentHashStr := hex.EncodeToString(bm.ParentHash[:])
		fmt.Fprintf(sb, "    \"%s\" -> \"%s\";\n", hashStr, parentHashStr)
	}
}

// getNodeColor returns the appropriate fill color based on block type.
func getNodeColor(isMainChain, isOrphan bool) string {
	if isOrphan {
		return colorOrphan
	} else if isMainChain {
		return colorMainChain
	}
	return colorSideChain
}

// shortenHash shortens a hash string to the first shortHashLength characters.
func shortenHash(hashStr string) string {
	if len(hashStr) > shortHashLength {
		return hashStr[:shortHashLength]
	}
	return hashStr
}

// formatBlockTooltip formats the block's transactions for display in a tooltip.
// Format:
//
//	blockHeaderHash
//	inputs   |outputs
//	txID (first transaction - coinbase)
//	0000(0)  |PubKeyHash(value)
//	--------------
//	...
func formatBlockTooltip(blk *block.Block) string {
	var sb strings.Builder

	hash := blk.Hash()
	blockHash := hex.EncodeToString(hash[:])
	sb.WriteString(shortenHash(blockHash))
	sb.WriteString("\\n\\n")
	sb.WriteString("inputs   |outputs\\n")

	for i, tx := range blk.Transactions {
		formatTransaction(&sb, &tx, i == 0)
		if i < len(blk.Transactions)-1 {
			sb.WriteString("--------------\\n")
		}
	}

	return sb.String()
}

// formatTransaction formats a single transaction for tooltip display.
// isCoinbase indicates if this is the first (coinbase) transaction in a block.
func formatTransaction(sb *strings.Builder, tx *transaction.Transaction, isCoinbase bool) {
	// Transaction ID
	txID := tx.Hash()
	txIDStr := hex.EncodeToString(txID[:])
	sb.WriteString(shortenHash(txIDStr))
	if isCoinbase {
		sb.WriteString(" (coinbase)")
	}
	sb.WriteString("\\n")

	// Determine max rows needed (max of inputs or outputs)
	maxRows := len(tx.Inputs)
	if len(tx.Outputs) > maxRows {
		maxRows = len(tx.Outputs)
	}

	for row := 0; row < maxRows; row++ {
		// Format input side
		if row < len(tx.Inputs) {
			input := &tx.Inputs[row]
			prevTxIDStr := hex.EncodeToString(input.PrevTxID[:])
			sb.WriteString(fmt.Sprintf("%s(%d)", shortenHash(prevTxIDStr), input.OutputIndex))
		} else {
			sb.WriteString("         ")
		}

		sb.WriteString("  |")

		// Format output side
		if row < len(tx.Outputs) {
			output := &tx.Outputs[row]
			pubKeyHashStr := hex.EncodeToString(output.PubKeyHash[:])
			sb.WriteString(fmt.Sprintf("%s(%d)", shortenHash(pubKeyHashStr), output.Value))
		}

		sb.WriteString("\\n")
	}
}
