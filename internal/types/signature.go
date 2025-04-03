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

func (s *Signature) V(chainID *big.Int) *big.Int {
	oddYParity := uint64(0)
	if s.OddYParity {
		oddYParity = 1
	}
	return new(big.Int).SetUint64(oddYParity)
}

func (s *Signature) LegacyV(chainID *big.Int) *big.Int {
	oddYParity := uint64(0)
	if s.OddYParity {
		oddYParity = 1
	}
	return new(big.Int).SetUint64(oddYParity + 35 + 2*chainID.Uint64())
}

type signatureMarshaling struct {
	R *math.HexOrDecimal256 `json:"r" gencodec:"required"`
	S *math.HexOrDecimal256 `json:"s" gencodec:"required"`
}
