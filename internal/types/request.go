package types

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
)

type Request struct {
	ptr unsafe.Pointer
}

func (t *Request) UnmarshalJSON(input []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(input, &raw); err != nil {
		return err
	}
	for key, val := range raw {
		switch key {
		case "DepositRequest":
			var req types.Deposit
			if err := json.Unmarshal(val, &req); err != nil {
				return err
			}
			t.ptr = unsafe.Pointer(&req)
		case "WithdrawalRequest":
			var req WithdrawalRequest
			if err := json.Unmarshal(val, &req); err != nil {
				return err
			}
			t.ptr = unsafe.Pointer(&req)
		case "ConsolidationRequest":
			var req ConsolidationRequest
			if err := json.Unmarshal(val, &req); err != nil {
				return err
			}
			t.ptr = unsafe.Pointer(&req)
		default:
			return fmt.Errorf("unknown request type: %s", key)
		}
	}
	return nil
}

//go:generate go run github.com/fjl/gencodec -type WithdrawalRequest -field-override withdrawalRequestMarshaling -out gen_withdrawal_request.go
type WithdrawalRequest struct {
	SourceAddress   common.Address `json:"sourceAddress" gencodec:"required"`
	ValidatorPubkey [48]byte       `json:"validatorPubkey" gencodec:"required"`
	Amount          uint64         `json:"amount" gencodec:"required"`
}

type withdrawalRequestMarshaling struct {
	ValidatorPubkey hexutil.Bytes       `json:"validatorPubkey" gencodec:"required"`
	Amount          math.HexOrDecimal64 `json:"amount" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type ConsolidationRequest -field-override consolidationRequestMarshaling -out gen_consolidation_request.go
type ConsolidationRequest struct {
	SourceAddress common.Address `json:"sourceAddress" gencodec:"required"`
	SourcePubkey  [48]byte       `json:"sourcePubkey" gencodec:"required"`
	TargetPubkey  [48]byte       `json:"targetPubkey" gencodec:"required"`
}

type consolidationRequestMarshaling struct {
	SourcePubkey hexutil.Bytes `json:"sourcePubkey" gencodec:"required"`
	TargetPubkey hexutil.Bytes `json:"targetPubkey" gencodec:"required"`
}
