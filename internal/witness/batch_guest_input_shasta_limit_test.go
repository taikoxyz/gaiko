package witness

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/manifest"
)

func TestDecodeShastaDerivationSourceManifest_RejectsTooManyBlocks(t *testing.T) {
	blocks := make([]*manifest.BlockManifest, shastaProposalMaxBlocks+1)
	for i := range blocks {
		blocks[i] = &manifest.BlockManifest{
			Timestamp:         uint64(i + 1),
			Coinbase:          common.Address{0x01},
			AnchorBlockNumber: 1,
			GasLimit:          1,
			Transactions:      types.Transactions{},
		}
	}

	encoded, err := rlp.EncodeToBytes(&legacyDerivationSourceManifest{Blocks: blocks})
	require.NoError(t, err)

	_, err = decodeShastaDerivationSourceManifest(encoded)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds limit")
}
