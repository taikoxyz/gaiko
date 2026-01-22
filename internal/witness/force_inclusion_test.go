package witness

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/manifest"
)

func TestValidateForceIncProposalManifestAllowsNonZeroFields(t *testing.T) {
	m := &manifest.DerivationSourceManifest{
		Blocks: []*manifest.BlockManifest{
			{
				Timestamp:         123,
				Coinbase:          common.HexToAddress("0x0000000000000000000000000000000000000001"),
				AnchorBlockNumber: 456,
				GasLimit:          789,
				Transactions:      types.Transactions{},
			},
		},
	}

	if !validateForceIncProposalManifest(m) {
		t.Fatalf("expected force inclusion manifest with non-zero fields to be valid")
	}
}
