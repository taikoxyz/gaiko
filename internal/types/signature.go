package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
)

//go:generate go run github.com/fjl/gencodec -type Signature -field-override signatureMarshaling -out gen_signature.go
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
	if chainID == nil {
		// self.odd_y_parity as u64 + chain_id * 2 + 35
		return new(big.Int).SetUint64(oddYParity + 35 + chainID.Uint64()*2)
	}
	return new(big.Int).SetUint64(oddYParity)
}

type signatureMarshaling struct {
	R *math.HexOrDecimal256 `json:"r" gencodec:"required"`
	S *math.HexOrDecimal256 `json:"s" gencodec:"required"`
}
