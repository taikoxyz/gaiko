package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
)

//go:generate go run github.com/fjl/gencodec -type Signature -field-override signatureMarshaling -out gen_signature.go

// Signature represents a signature of a transaction has the same format with raiko.
type Signature struct {
	R          *big.Int `json:"r"            gencodec:"required"`
	S          *big.Int `json:"s"            gencodec:"required"`
	OddYParity bool     `json:"odd_y_parity" gencodec:"required"`
}

func (s *Signature) V(chainID *big.Int, isLegacy bool) *big.Int {
	oddYParity := uint64(0)
	if s.OddYParity {
		oddYParity = 1
	}
	if isLegacy {
		if chainID == nil {
			// non-EIP-155 legacy scheme, v = 27 for even y-parity, v = 28 for odd y-parity
			if s.OddYParity {
				return new(big.Int).SetUint64(28)
			}
			return new(big.Int).SetUint64(27)
		}
		// EIP-155: v = {0, 1} + CHAIN_ID * 2 + 35
		return new(big.Int).SetUint64(oddYParity + 35 + chainID.Uint64()*2)
	}
	return new(big.Int).SetUint64(oddYParity)
}

type signatureMarshaling struct {
	R *math.HexOrDecimal256 `json:"r" gencodec:"required"`
	S *math.HexOrDecimal256 `json:"s" gencodec:"required"`
}
