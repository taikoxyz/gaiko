package transition

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var emptyHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

const (
	anchorGasLimit      uint32 = 250_000
	blobBytes                  = params.BlobTxBytesPerFieldElement * params.BlobTxFieldElementsPerBlob
	blockMaxTxListBytes uint64 = (params.BlobTxBytesPerFieldElement - 1) * params.BlobTxFieldElementsPerBlob
	blockMaxGasLimit           = 240_000_000
)
