package transition

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	txListDecompressor "github.com/taikoxyz/taiko-mono/packages/taiko-client/driver/txlist_decompressor"
)

func decompressTxList(
	txListBytes []byte,
	blobUsed, isPacaya bool,
	chainID *big.Int,
) types.Transactions {
	return txListDecompressor.NewTxListDecompressor(
		blockMaxGasLimit,
		blockMaxTxListBytes,
		chainID,
	).TryDecompress(
		chainID,
		txListBytes,
		blobUsed,
		isPacaya,
	)
}

// sliceTxList returns the sliced txList bytes from the given offset and length.
func sliceTxList(id *big.Int, b []byte, offset, length uint32) ([]byte, error) {
	if offset+length > uint32(len(b)) {
		return nil, fmt.Errorf(
			"invalid txlist offset and size in metadata (%d): offset=%d, size=%d, blobSize=%d", id, offset, length, len(b),
		)
	}
	return b[offset : offset+length], nil
}
