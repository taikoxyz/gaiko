package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

//go:generate go run github.com/fjl/gencodec -type BlockProposedV2 -field-override blockProposedV2Marshaling -out gen_block_proposed_v2.go

// BlockProposedV2 represents a BlockProposed event raised by the TaikoL1Client contract.
type BlockProposedV2 struct {
	BlockId *big.Int         `json:"blockId" gencodec:"required"`
	Meta    *BlockMetadataV2 `json:"meta" gencodec:"required"`
}

func (b *BlockProposedV2) Origin() *ontake.TaikoL1ClientBlockProposedV2 {
	return &ontake.TaikoL1ClientBlockProposedV2{
		BlockId: b.BlockId,
		Meta: ontake.TaikoDataBlockMetadataV2{
			AnchorBlockHash:  b.Meta.AnchorBlockHash,
			Difficulty:       b.Meta.Difficulty,
			BlobHash:         b.Meta.BlobHash,
			ExtraData:        b.Meta.ExtraData,
			Coinbase:         b.Meta.Coinbase,
			Id:               b.Meta.Id,
			GasLimit:         b.Meta.GasLimit,
			Timestamp:        b.Meta.Timestamp,
			AnchorBlockId:    b.Meta.AnchorBlockId,
			MinTier:          b.Meta.MinTier,
			BlobUsed:         b.Meta.BlobUsed,
			ParentMetaHash:   b.Meta.ParentMetaHash,
			Proposer:         b.Meta.Proposer,
			LivenessBond:     b.Meta.LivenessBond,
			ProposedAt:       b.Meta.ProposedAt,
			ProposedIn:       b.Meta.ProposedIn,
			BlobTxListOffset: b.Meta.BlobTxListOffset,
			BlobTxListLength: b.Meta.BlobTxListLength,
			BlobIndex:        b.Meta.BlobIndex,
			BaseFeeConfig: ontake.LibSharedDataBaseFeeConfig{
				AdjustmentQuotient:     b.Meta.BaseFeeConfig.AdjustmentQuotient,
				SharingPctg:            b.Meta.BaseFeeConfig.SharingPctg,
				GasIssuancePerSecond:   b.Meta.BaseFeeConfig.GasIssuancePerSecond,
				MinGasExcess:           b.Meta.BaseFeeConfig.MinGasExcess,
				MaxGasIssuancePerBlock: b.Meta.BaseFeeConfig.MaxGasIssuancePerBlock,
			},
		},
	}
}

type blockProposedV2Marshaling struct {
	BlockId *math.HexOrDecimal256 `json:"blockId" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BlockMetadataV2 -field-override blockMetadataV2Marshaling -out gen_block_metadata_v2.go

// BlockMetadataV2 is an auto generated low-level Go binding around an user-defined struct.
type BlockMetadataV2 struct {
	AnchorBlockHash  common.Hash                 `json:"anchorBlockHash" gencodec:"required"`
	Difficulty       common.Hash                 `json:"difficulty" gencodec:"required"`
	BlobHash         common.Hash                 `json:"blobHash" gencodec:"required"`
	ExtraData        common.Hash                 `json:"extraData" gencodec:"required"`
	Coinbase         common.Address              `json:"coinbase" gencodec:"required"`
	Id               uint64                      `json:"id" gencodec:"required"`
	GasLimit         uint32                      `json:"gasLimit" gencodec:"required"`
	Timestamp        uint64                      `json:"timestamp" gencodec:"required"`
	AnchorBlockId    uint64                      `json:"anchorBlockId" gencodec:"required"`
	MinTier          uint16                      `json:"minTier" gencodec:"required"`
	BlobUsed         bool                        `json:"blobUsed" gencodec:"required"`
	ParentMetaHash   common.Hash                 `json:"parentMetaHash" gencodec:"required"`
	Proposer         common.Address              `json:"proposer" gencodec:"required"`
	LivenessBond     *big.Int                    `json:"livenessBond" gencodec:"required"`
	ProposedAt       uint64                      `json:"proposedAt" gencodec:"required"`
	ProposedIn       uint64                      `json:"proposedIn" gencodec:"required"`
	BlobTxListOffset uint32                      `json:"blobTxListOffset" gencodec:"required"`
	BlobTxListLength uint32                      `json:"blobTxListLength" gencodec:"required"`
	BlobIndex        uint8                       `json:"blobIndex" gencodec:"required"`
	BaseFeeConfig    *LibSharedDataBaseFeeConfig `json:"baseFeeConfig" gencodec:"required"`
}

type blockMetadataV2Marshaling struct {
	Id            math.HexOrDecimal64   `json:"id" gencodec:"required"`
	Timestamp     math.HexOrDecimal64   `json:"timestamp" gencodec:"required"`
	AnchorBlockId math.HexOrDecimal64   `json:"anchorBlockId" gencodec:"required"`
	LivenessBond  *math.HexOrDecimal256 `json:"livenessBond" gencodec:"required"`
	ProposedAt    math.HexOrDecimal64   `json:"proposedAt" gencodec:"required"`
	ProposedIn    math.HexOrDecimal64   `json:"proposedIn" gencodec:"required"`
}
