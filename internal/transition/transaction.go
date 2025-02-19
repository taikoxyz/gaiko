package transition

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	txListDecompressor "github.com/taikoxyz/taiko-mono/packages/taiko-client/driver/txlist_decompressor"
	// "github.com/taikoxyz/taiko-mono/packages/taiko-client/pkg/rpc"
)

func decodeTxs(
	txListBytes []byte,
	blobUsed, isPacaya bool,
	chainID, blockNumber *big.Int,
	offset, length uint32) types.Transactions {
	BlockMaxTxListBytes := uint64(100000000)
	decompressor := txListDecompressor.NewTxListDecompressor(params.MaxGasLimit, BlockMaxTxListBytes, chainID)
	if blobUsed {
		blob := eth.Blob(txListBytes)
		var err error
		if txListBytes, err = blob.ToData(); err != nil {
			return nil
		}
		if txListBytes, err = sliceTxList(blockNumber, txListBytes, offset, length); err != nil {
			return nil
		}
	}
	return decompressor.TryDecompress(chainID, txListBytes, blobUsed, isPacaya)
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
