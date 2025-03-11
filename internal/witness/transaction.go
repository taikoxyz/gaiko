package witness

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	txListDecompressor "github.com/taikoxyz/taiko-mono/packages/taiko-client/driver/txlist_decompressor"
)

func decompressTxList(
	txListBytes []byte,
	maxBytesPerTxList uint64,
	blobUsed, isPacaya bool,
	chainID *big.Int,
) types.Transactions {
	return txListDecompressor.NewTxListDecompressor(
		blockMaxGasLimit,
		maxBytesPerTxList,
		chainID,
	).TryDecompress(
		chainID,
		txListBytes,
		blobUsed,
		isPacaya,
	)
}

// sliceTxList returns the sliced txList bytes from the given offset and length.
func sliceTxList(id *big.Int, b []byte, slice *Slice) ([]byte, error) {
	if slice == nil {
		return b, nil
	}
	if slice.Offset+slice.Length > uint32(len(b)) {
		return nil, fmt.Errorf(
			"invalid txlist offset and size in metadata (%d): offset=%d, size=%d, blobSize=%d",
			id, slice.Offset, slice.Length, len(b),
		)
	}
	return b[slice.Offset : slice.Offset+slice.Length], nil
}
