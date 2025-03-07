// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
)

var _ = (*txEip2930Marshaling)(nil)

// MarshalJSON marshals as JSON.
func (t TxEip2930) MarshalJSON() ([]byte, error) {
	type TxEip2930 struct {
		ChainID    *math.HexOrDecimal256 `json:"chain_id"    gencodec:"required"`
		Nonce      math.HexOrDecimal64   `json:"nonce"       gencodec:"required"`
		GasPrice   *math.HexOrDecimal256 `json:"gas_price"   gencodec:"required"`
		GasLimit   math.HexOrDecimal64   `json:"gas_limit"   gencodec:"required"`
		To         *common.Address       `json:"to"`
		Value      *math.HexOrDecimal256 `json:"value"       gencodec:"required"`
		AccessList AccessList            `json:"access_list" gencodec:"required"`
		Input      hexutil.Bytes         `json:"input"       gencodec:"required"`
	}
	var enc TxEip2930
	enc.ChainID = (*math.HexOrDecimal256)(t.ChainID)
	enc.Nonce = math.HexOrDecimal64(t.Nonce)
	enc.GasPrice = (*math.HexOrDecimal256)(t.GasPrice)
	enc.GasLimit = math.HexOrDecimal64(t.GasLimit)
	enc.To = t.To
	enc.Value = (*math.HexOrDecimal256)(t.Value)
	enc.AccessList = t.AccessList
	enc.Input = t.Input
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (t *TxEip2930) UnmarshalJSON(input []byte) error {
	type TxEip2930 struct {
		ChainID    *math.HexOrDecimal256 `json:"chain_id"    gencodec:"required"`
		Nonce      *math.HexOrDecimal64  `json:"nonce"       gencodec:"required"`
		GasPrice   *math.HexOrDecimal256 `json:"gas_price"   gencodec:"required"`
		GasLimit   *math.HexOrDecimal64  `json:"gas_limit"   gencodec:"required"`
		To         *common.Address       `json:"to"`
		Value      *math.HexOrDecimal256 `json:"value"       gencodec:"required"`
		AccessList *AccessList           `json:"access_list" gencodec:"required"`
		Input      *hexutil.Bytes        `json:"input"       gencodec:"required"`
	}
	var dec TxEip2930
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.ChainID == nil {
		return errors.New("missing required field 'chain_id' for TxEip2930")
	}
	t.ChainID = (*big.Int)(dec.ChainID)
	if dec.Nonce == nil {
		return errors.New("missing required field 'nonce' for TxEip2930")
	}
	t.Nonce = uint64(*dec.Nonce)
	if dec.GasPrice == nil {
		return errors.New("missing required field 'gas_price' for TxEip2930")
	}
	t.GasPrice = (*big.Int)(dec.GasPrice)
	if dec.GasLimit == nil {
		return errors.New("missing required field 'gas_limit' for TxEip2930")
	}
	t.GasLimit = uint64(*dec.GasLimit)
	if dec.To != nil {
		t.To = dec.To
	}
	if dec.Value == nil {
		return errors.New("missing required field 'value' for TxEip2930")
	}
	t.Value = (*big.Int)(dec.Value)
	if dec.AccessList == nil {
		return errors.New("missing required field 'access_list' for TxEip2930")
	}
	t.AccessList = *dec.AccessList
	if dec.Input == nil {
		return errors.New("missing required field 'input' for TxEip2930")
	}
	t.Input = *dec.Input
	return nil
}
