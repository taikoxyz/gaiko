package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

type TransactionSignedList []*TransactionSigned

func (t TransactionSignedList) GethType() []*types.Transaction {
	if t == nil {
		log.Warn("missing TransactionSignedList when converting to GethType")
		return nil
	}
	txs := make([]*types.Transaction, len(t))
	for i, tx := range t {
		txs[i] = tx.GethType()
	}
	return txs
}

type TransactionSigned struct {
	Hash        common.Hash  `json:"hash"        gencodec:"required"`
	Signature   *Signature   `json:"signature"   gencodec:"required"`
	Transaction *Transaction `json:"transaction" gencodec:"required"`
}

type Transaction struct {
	inner any
}

func (t *TransactionSigned) GethType() *types.Transaction {
	if t == nil {
		log.Warn("missing TransactionSigned when converting to GethType")
		return nil
	}
	switch inner := t.Transaction.inner.(type) {
	case *TxLegacy:
		tx := &types.LegacyTx{
			Nonce:    inner.Nonce,
			GasPrice: inner.GasPrice,
			Gas:      inner.GasLimit,
			To:       inner.To,
			Value:    inner.Value,
			Data:     inner.Input,
			V:        t.Signature.V(inner.ChainID, true),
			R:        t.Signature.R,
			S:        t.Signature.S,
		}
		return types.NewTx(tx)
	case *TxEip2930:
		tx := &types.AccessListTx{
			ChainID:    inner.ChainID,
			Nonce:      inner.Nonce,
			GasPrice:   inner.GasPrice,
			Gas:        inner.GasLimit,
			To:         inner.To,
			Value:      inner.Value,
			Data:       inner.Input,
			AccessList: inner.AccessList.GethType(),
			V:          t.Signature.V(inner.ChainID, false),
			R:          t.Signature.R,
			S:          t.Signature.S,
		}
		return types.NewTx(tx)
	case *TxEip1559:
		tx := &types.DynamicFeeTx{
			ChainID:    inner.ChainID,
			Nonce:      inner.Nonce,
			GasTipCap:  inner.MaxPriorityFeePerGas,
			GasFeeCap:  inner.MaxFeePerGas,
			Gas:        inner.GasLimit,
			To:         inner.To,
			Value:      inner.Value,
			Data:       inner.Input,
			AccessList: inner.AccessList.GethType(),
			V:          t.Signature.V(inner.ChainID, false),
			R:          t.Signature.R,
			S:          t.Signature.S,
		}
		return types.NewTx(tx)
	case *TxEip4844:
		tx := &types.BlobTx{
			ChainID:    uint256.MustFromBig(inner.ChainID),
			Nonce:      inner.Nonce,
			GasTipCap:  uint256.MustFromBig(inner.MaxPriorityFeePerGas),
			GasFeeCap:  uint256.MustFromBig(inner.MaxFeePerGas),
			Gas:        inner.GasLimit,
			To:         inner.To,
			Value:      uint256.MustFromBig(inner.Value),
			Data:       inner.Input,
			AccessList: inner.AccessList.GethType(),
			BlobFeeCap: uint256.MustFromBig(inner.MaxFeePerBlobGas),
			BlobHashes: inner.BlobVersionedHashes,
			V:          uint256.MustFromBig(t.Signature.V(inner.ChainID, false)),
			R:          uint256.MustFromBig(t.Signature.R),
			S:          uint256.MustFromBig(t.Signature.S),
		}
		return types.NewTx(tx)
	default:
		panic(fmt.Sprintf("unknown transaction type: %T", inner))
	}
}

func (t *Transaction) UnmarshalJSON(input []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(input, &raw); err != nil {
		return err
	}
	for key, val := range raw {
		switch key {
		case "Legacy":
			var inner TxLegacy
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			t.inner = &inner
		case "Eip2930":
			var inner TxEip2930
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			t.inner = &inner
		case "Eip1559":
			var inner TxEip1559
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			t.inner = &inner
		case "Eip4844":
			var inner TxEip4844
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			t.inner = &inner
		default:
			return fmt.Errorf("unknown transaction type: %s", key)
		}
	}
	return nil
}

//go:generate go run github.com/fjl/gencodec -type TxLegacy -field-override txLegacyMarshaling -out gen_tx_legacy.go

type TxLegacy struct {
	ChainID  *big.Int        `json:"chain_id"`
	Nonce    uint64          `json:"nonce"     gencodec:"required"`
	GasPrice *big.Int        `json:"gas_price" gencodec:"required"`
	GasLimit uint64          `json:"gas_limit" gencodec:"required"`
	To       *common.Address `json:"to"`
	Value    *big.Int        `json:"value"     gencodec:"required"`
	Input    []byte          `json:"input"     gencodec:"required"`
}

type txLegacyMarshaling struct {
	ChainID  *math.HexOrDecimal256 `json:"chain_id"`
	Nonce    math.HexOrDecimal64   `json:"nonce"     gencodec:"required"`
	GasPrice *math.HexOrDecimal256 `json:"gas_price" gencodec:"required"`
	GasLimit math.HexOrDecimal64   `json:"gas_limit" gencodec:"required"`
	Value    *math.HexOrDecimal256 `json:"value"     gencodec:"required"`
	Input    hexutil.Bytes         `json:"input"     gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type TxEip2930 -field-override txEip2930Marshaling -out gen_tx_eip2930.go

type TxEip2930 struct {
	ChainID    *big.Int        `json:"chain_id"    gencodec:"required"`
	Nonce      uint64          `json:"nonce"       gencodec:"required"`
	GasPrice   *big.Int        `json:"gas_price"   gencodec:"required"`
	GasLimit   uint64          `json:"gas_limit"   gencodec:"required"`
	To         *common.Address `json:"to"`
	Value      *big.Int        `json:"value"       gencodec:"required"`
	AccessList AccessList      `json:"access_list" gencodec:"required"`
	Input      []byte          `json:"input"       gencodec:"required"`
}

type txEip2930Marshaling struct {
	ChainID  *math.HexOrDecimal256 `json:"chain_id"  gencodec:"required"`
	Nonce    math.HexOrDecimal64   `json:"nonce"     gencodec:"required"`
	GasPrice *math.HexOrDecimal256 `json:"gas_price" gencodec:"required"`
	GasLimit math.HexOrDecimal64   `json:"gas_limit" gencodec:"required"`
	Value    *math.HexOrDecimal256 `json:"value"     gencodec:"required"`
	Input    hexutil.Bytes         `json:"input"     gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type TxEip1559 -field-override txEip1559Marshaling -out gen_tx_eip1559.go

type TxEip1559 struct {
	ChainID              *big.Int        `json:"chain_id"                 gencodec:"required"`
	Nonce                uint64          `json:"nonce"                    gencodec:"required"`
	GasLimit             uint64          `json:"gas_limit"                gencodec:"required"`
	MaxFeePerGas         *big.Int        `json:"max_fee_per_gas"          gencodec:"required"`
	MaxPriorityFeePerGas *big.Int        `json:"max_priority_fee_per_gas" gencodec:"required"`
	To                   *common.Address `json:"to"`
	Value                *big.Int        `json:"value"                    gencodec:"required"`
	AccessList           AccessList      `json:"access_list"`
	Input                []byte          `json:"input"                    gencodec:"required"`
}

type txEip1559Marshaling struct {
	ChainID              *math.HexOrDecimal256 `json:"chain_id"                 gencodec:"required"`
	Nonce                math.HexOrDecimal64   `json:"nonce"                    gencodec:"required"`
	GasLimit             math.HexOrDecimal64   `json:"gas_limit"                gencodec:"required"`
	MaxFeePerGas         *math.HexOrDecimal256 `json:"max_fee_per_gas"          gencodec:"required"`
	MaxPriorityFeePerGas *math.HexOrDecimal256 `json:"max_priority_fee_per_gas" gencodec:"required"`
	Value                *math.HexOrDecimal256 `json:"value"                    gencodec:"required"`
	Input                hexutil.Bytes         `json:"input"                    gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type TxEip4844 -field-override txEip4844Marshaling -out gen_tx_eip4844.go

// TODO: unsupported in l2
type TxEip4844 struct {
	ChainID              *big.Int       `json:"chain_id"                 gencodec:"required"`
	Nonce                uint64         `json:"nonce"                    gencodec:"required"`
	GasLimit             uint64         `json:"gas_limit"                gencodec:"required"`
	MaxFeePerGas         *big.Int       `json:"max_fee_per_gas"          gencodec:"required"`
	MaxPriorityFeePerGas *big.Int       `json:"max_priority_fee_per_gas" gencodec:"required"`
	To                   common.Address `json:"to"                       gencodec:"required"`
	Value                *big.Int       `json:"value"                    gencodec:"required"`
	AccessList           AccessList     `json:"access_list"              gencodec:"required"`
	BlobVersionedHashes  []common.Hash  `json:"blob_versioned_hashes"    gencodec:"required"`
	MaxFeePerBlobGas     *big.Int       `json:"max_fee_per_blob_gas"     gencodec:"required"`
	Input                []byte         `json:"input"                    gencodec:"required"`
}

type txEip4844Marshaling struct {
	ChainID              *math.HexOrDecimal256 `json:"chain_id"                 gencodec:"required"`
	Nonce                math.HexOrDecimal64   `json:"nonce"                    gencodec:"required"`
	GasLimit             math.HexOrDecimal64   `json:"gas_limit"                gencodec:"required"`
	MaxFeePerGas         *math.HexOrDecimal256 `json:"max_fee_per_gas"          gencodec:"required"`
	MaxPriorityFeePerGas *math.HexOrDecimal256 `json:"max_priority_fee_per_gas" gencodec:"required"`
	Value                *math.HexOrDecimal256 `json:"value"                    gencodec:"required"`
	MaxFeePerBlobGas     *math.HexOrDecimal256 `json:"max_fee_per_blob_gas"     gencodec:"required"`
	Input                hexutil.Bytes         `json:"input"                    gencodec:"required"`
}
