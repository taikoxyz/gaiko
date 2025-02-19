package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

//go:generate go run github.com/fjl/gencodec -type BlockProposed -field-override blockProposedMarshaling -out gen_block_proposed.go

// BlockProposed represents a BlockProposed event raised by the TaikoL1Client contract.
type BlockProposed struct {
	BlockId           *big.Int       `json:"blockId" gencodec:"required"`
	AssignedProver    common.Address `json:"assignedProver" gencodec:"required"`
	LivenessBond      *big.Int       `json:"livenessBond" gencodec:"required"`
	Meta              *BlockMetadata `json:"meta" gencodec:"required"`
	DepositsProcessed []*EthDeposit  `json:"depositsProcessed" gencodec:"required"`
}

func (b *BlockProposed) GethType() *ontake.TaikoL1ClientBlockProposed {
	deposits := make([]ontake.TaikoDataEthDeposit, len(b.DepositsProcessed))
	for i, deposit := range b.DepositsProcessed {
		deposits[i] = ontake.TaikoDataEthDeposit{
			Recipient: deposit.Recipient,
			Amount:    deposit.Amount,
			Id:        deposit.Id,
		}
	}
	return &ontake.TaikoL1ClientBlockProposed{
		BlockId:        b.BlockId,
		AssignedProver: b.AssignedProver,
		LivenessBond:   b.LivenessBond,
		Meta: ontake.TaikoDataBlockMetadata{
			L1Hash:         b.Meta.L1Hash,
			Difficulty:     b.Meta.Difficulty,
			BlobHash:       b.Meta.BlobHash,
			ExtraData:      b.Meta.ExtraData,
			DepositsHash:   b.Meta.DepositsHash,
			Coinbase:       b.Meta.Coinbase,
			Id:             b.Meta.Id,
			GasLimit:       b.Meta.GasLimit,
			Timestamp:      b.Meta.Timestamp,
			L1Height:       b.Meta.L1Height,
			MinTier:        b.Meta.MinTier,
			BlobUsed:       b.Meta.BlobUsed,
			ParentMetaHash: b.Meta.ParentMetaHash,
			Sender:         b.Meta.Sender,
		},
		DepositsProcessed: deposits,
	}
}

type blockProposedMarshaling struct {
	BlockId      *math.HexOrDecimal256 `json:"blockId" gencodec:"required"`
	LivenessBond *math.HexOrDecimal256 `json:"livenessBond" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BlockMetadata -field-override blockMetadataMarshaling -out gen_block_metadata.go

// BlockMetadata is an auto generated low-level Go binding around an user-defined struct.
type BlockMetadata struct {
	L1Hash         common.Hash    `json:"l1Hash" gencodec:"required"`
	Difficulty     common.Hash    `json:"difficulty" gencodec:"required"`
	BlobHash       common.Hash    `json:"blobHash" gencodec:"required"`
	ExtraData      common.Hash    `json:"extraData" gencodec:"required"`
	DepositsHash   common.Hash    `json:"depositsHash" gencodec:"required"`
	Coinbase       common.Address `json:"coinbase" gencodec:"required"`
	Id             uint64         `json:"id" gencodec:"required"`
	GasLimit       uint32         `json:"gasLimit" gencodec:"required"`
	Timestamp      uint64         `json:"timestamp" gencodec:"required"`
	L1Height       uint64         `json:"l1Height" gencodec:"required"`
	MinTier        uint16         `json:"minTier" gencodec:"required"`
	BlobUsed       bool           `json:"blobUsed" gencodec:"required"`
	ParentMetaHash common.Hash    `json:"parentMetaHash" gencodec:"required"`
	Sender         common.Address `json:"sender" gencodec:"required"`
}

type blockMetadataMarshaling struct {
	Id        math.HexOrDecimal64 `json:"id" gencodec:"required"`
	Timestamp math.HexOrDecimal64 `json:"timestamp" gencodec:"required"`
	L1Height  math.HexOrDecimal64 `json:"l1Height" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type EthDeposit -field-override ethDepositMarshaling -out gen_eth_deposit.go

// EthDeposit is an auto generated low-level Go binding around an user-defined struct.
type EthDeposit struct {
	Recipient common.Address `json:"recipient" gencodec:"required"`
	Amount    *big.Int       `json:"amount" gencodec:"required"`
	Id        uint64         `json:"id" gencodec:"required"`
}

type ethDepositMarshaling struct {
	Amount *math.HexOrDecimal256 `json:"amount" gencodec:"required"`
	Id     math.HexOrDecimal64   `json:"id" gencodec:"required"`
}
