package types

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
)

type Requests []any

func (t *Requests) UnmarshalJSON(data []byte) error {
	raw := []map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, req := range raw {
		for key, val := range req {
			switch key {
			case "DepositRequest":
				var inner Deposit
				if err := json.Unmarshal(val, &inner); err != nil {
					return err
				}
				*t = append(*t, &inner)
			case "WithdrawalRequest":
				var inner WithdrawalRequest
				if err := json.Unmarshal(val, &inner); err != nil {
					return err
				}
				*t = append(*t, &inner)
			case "ConsolidationRequest":
				var inner ConsolidationRequest
				if err := json.Unmarshal(val, &inner); err != nil {
					return err
				}
				*t = append(*t, &inner)
			default:
				return fmt.Errorf("unknown request type: %s", key)
			}
		}
	}
	return nil
}

func (r Requests) GethType() []*types.Request {
	requests := make([]*types.Request, len(r))
	for i, req := range r {
		switch inner := req.(type) {
		case *types.Deposit:
			requests[i] = types.NewRequest(inner)
		case *WithdrawalRequest:
			// TODO: not support yet
		case *ConsolidationRequest:
			// TODO: not support yet
		default:
			panic(fmt.Sprintf("unknown request type: %T", inner))
		}
	}
	return requests
}

//go:generate go run github.com/fjl/gencodec -type Deposit -field-override depositMarshaling -out gen_deposit.go

type Deposit struct {
	PublicKey             [48]byte    `json:"pubkey"`                 // public key of validator
	WithdrawalCredentials common.Hash `json:"withdrawal_credentials"` // beneficiary of the validator funds
	Amount                uint64      `json:"amount"`                 // deposit size in Gwei
	Signature             [96]byte    `json:"signature"`              // signature over deposit msg
	Index                 uint64      `json:"index"`                  // deposit count value
}

// field type overrides for gencodec
type depositMarshaling struct {
	PublicKey             hexutil.Bytes
	WithdrawalCredentials hexutil.Bytes
	Amount                hexutil.Uint64
	Signature             hexutil.Bytes
	Index                 hexutil.Uint64
}

//go:generate go run github.com/fjl/gencodec -type WithdrawalRequest -field-override withdrawalRequestMarshaling -out gen_withdrawal_request.go

type WithdrawalRequest struct {
	SourceAddress   common.Address `json:"source_address"   gencodec:"required"`
	ValidatorPubkey [48]byte       `json:"validator_pubkey" gencodec:"required"`
	Amount          uint64         `json:"amount"           gencodec:"required"`
}

type withdrawalRequestMarshaling struct {
	ValidatorPubkey hexutil.Bytes       `json:"validator_pubkey" gencodec:"required"`
	Amount          math.HexOrDecimal64 `json:"amount"           gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type ConsolidationRequest -field-override consolidationRequestMarshaling -out gen_consolidation_request.go

type ConsolidationRequest struct {
	SourceAddress common.Address `json:"source_address" gencodec:"required"`
	SourcePubkey  [48]byte       `json:"source_pubkey"  gencodec:"required"`
	TargetPubkey  [48]byte       `json:"target_pubkey"  gencodec:"required"`
}

type consolidationRequestMarshaling struct {
	SourcePubkey hexutil.Bytes `json:"source_pubkey" gencodec:"required"`
	TargetPubkey hexutil.Bytes `json:"target_pubkey" gencodec:"required"`
}
