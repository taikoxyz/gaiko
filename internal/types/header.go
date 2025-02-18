package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate go run github.com/fjl/gencodec -type Header -field-override headerMarshaling -out gen_header.go
type Header struct {
	ParentHash            common.Hash    `json:"parent_hash" gencodec:"required"`
	OmmerHash             common.Hash    `json:"ommers_hash" gencodec:"required"`
	Coinbase              common.Address `json:"beneficiary" gencodec:"required"`
	Root                  common.Hash    `json:"state_root" gencodec:"required"`
	TxHash                common.Hash    `json:"transactions_root" gencodec:"required"`
	ReceiptHash           common.Hash    `json:"receipts_root" gencodec:"required"`
	Bloom                 types.Bloom    `json:"logs_bloom" gencodec:"required"`
	Difficulty            *big.Int       `json:"difficulty" gencodec:"required"`
	Number                *big.Int       `json:"number" gencodec:"required"`
	GasLimit              uint64         `json:"gas_limit" gencodec:"required"`
	GasUsed               uint64         `json:"gas_used" gencodec:"required"`
	Time                  uint64         `json:"timestamp" gencodec:"required"`
	Extra                 []byte         `json:"extra_data" gencodec:"required"`
	MixDigest             common.Hash    `json:"mix_hash" gencodec:"required"`
	Nonce                 uint64         `json:"nonce" gencodec:"required"`
	BaseFee               *big.Int       `json:"base_fee_per_gas"`
	WithdrawalsHash       *common.Hash   `json:"withdrawals_root"`
	BlobGasUsed           *uint64        `json:"blob_gas_used"`
	ExcessBlobGas         *uint64        `json:"excess_blob_gas"`
	ParentBeaconBlockRoot *common.Hash   `json:"parent_beacon_block_root"`
	RequestsRoot          *common.Hash   `json:"requests_root"`
}

type headerMarshaling struct {
	Difficulty    *math.HexOrDecimal256 `json:"difficulty" gencodec:"required"`
	Number        *math.HexOrDecimal256 `json:"number" gencodec:"required"`
	GasLimit      math.HexOrDecimal64   `json:"gas_limit" gencodec:"required"`
	GasUsed       math.HexOrDecimal64   `json:"gas_used" gencodec:"required"`
	Time          math.HexOrDecimal64   `json:"timestamp" gencodec:"required"`
	Extra         hexutil.Bytes         `json:"extra_data" gencodec:"required"`
	Nonce         math.HexOrDecimal64   `json:"nonce" gencodec:"required"`
	BaseFee       *math.HexOrDecimal256 `json:"base_fee_per_gas"`
	BlobGasUsed   *math.HexOrDecimal64  `json:"blob_gas_used"`
	ExcessBlobGas *math.HexOrDecimal64  `json:"excess_blob_gas"`
}

func (h *Header) Origin() *types.Header {
	return &types.Header{
		ParentHash:       h.ParentHash,
		UncleHash:        h.OmmerHash,
		Coinbase:         h.Coinbase,
		Root:             h.Root,
		TxHash:           h.TxHash,
		ReceiptHash:      h.ReceiptHash,
		Bloom:            h.Bloom,
		Difficulty:       h.Difficulty,
		Number:           h.Number,
		GasLimit:         h.GasLimit,
		GasUsed:          h.GasUsed,
		Time:             h.Time,
		Extra:            h.Extra,
		MixDigest:        h.MixDigest,
		Nonce:            types.EncodeNonce(h.Nonce),
		BaseFee:          h.BaseFee,
		WithdrawalsHash:  h.WithdrawalsHash,
		BlobGasUsed:      h.BlobGasUsed,
		ExcessBlobGas:    h.ExcessBlobGas,
		ParentBeaconRoot: h.ParentBeaconBlockRoot,
		RequestsHash:     h.RequestsRoot,
	}
}
