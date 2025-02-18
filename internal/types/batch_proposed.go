package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

//go:generate go run github.com/fjl/gencodec -type BatchProposed -field-override batchProposedMarshaling -out gen_batch_proposed.go

// BatchProposed represents a BatchProposed event raised by the TaikoInboxClient contract.
type BatchProposed struct {
	Info   *BatchInfo     `json:"info" gencodec:"required"`
	Meta   *BatchMetadata `json:"meta" gencodec:"required"`
	TxList hexutil.Bytes  `json:"txList" gencodec:"required"`
}

func (b *BatchProposed) Origin() *pacaya.TaikoInboxClientBatchProposed {
	blocks := make([]pacaya.ITaikoInboxBlockParams, len(b.Info.Blocks))
	for i, block := range b.Info.Blocks {
		signalSlots := make([][32]byte, len(block.SignalSlots))
		for j, slot := range block.SignalSlots {
			signalSlots[j] = slot
		}
		blocks[i] = pacaya.ITaikoInboxBlockParams{
			NumTransactions: block.NumTransactions,
			TimeShift:       block.TimeShift,
			SignalSlots:     signalSlots,
		}
	}

	blobHashes := make([][32]byte, len(b.Info.BlobHashes))
	for i, hash := range b.Info.BlobHashes {
		blobHashes[i] = hash
	}
	return &pacaya.TaikoInboxClientBatchProposed{
		Info: pacaya.ITaikoInboxBatchInfo{
			TxsHash:            b.Info.TxsHash,
			Blocks:             blocks,
			BlobHashes:         blobHashes,
			ExtraData:          b.Info.ExtraData,
			Coinbase:           b.Info.Coinbase,
			ProposedIn:         b.Info.ProposedIn,
			BlobByteOffset:     b.Info.BlobByteOffset,
			BlobByteSize:       b.Info.BlobByteSize,
			GasLimit:           b.Info.GasLimit,
			LastBlockId:        b.Info.LastBlockId,
			LastBlockTimestamp: b.Info.LastBlockTimestamp,
			AnchorBlockId:      b.Info.AnchorBlockId,
			AnchorBlockHash:    b.Info.AnchorBlockHash,
			BaseFeeConfig: pacaya.LibSharedDataBaseFeeConfig{
				AdjustmentQuotient:     b.Info.BaseFeeConfig.AdjustmentQuotient,
				SharingPctg:            b.Info.BaseFeeConfig.SharingPctg,
				GasIssuancePerSecond:   b.Info.BaseFeeConfig.GasIssuancePerSecond,
				MinGasExcess:           b.Info.BaseFeeConfig.MinGasExcess,
				MaxGasIssuancePerBlock: b.Info.BaseFeeConfig.MaxGasIssuancePerBlock,
			},
		},
		Meta: pacaya.ITaikoInboxBatchMetadata{
			InfoHash:   b.Meta.InfoHash,
			Proposer:   b.Meta.Proposer,
			BatchId:    b.Meta.BatchId,
			ProposedAt: b.Meta.ProposedAt,
		},
		TxList: b.TxList,
	}
}

type batchProposedMarshaling struct {
	TxList hexutil.Bytes `json:"txList" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BatchMetadata -field-override batchMetadataMarshaling -out gen_batch_metadata.go

// BatchMetadata is an auto generated low-level Go binding around an user-defined struct.
type BatchMetadata struct {
	InfoHash   common.Hash    `json:"infoHash" gencodec:"required"`
	Proposer   common.Address `json:"proposer" gencodec:"required"`
	BatchId    uint64         `json:"batchId" gencodec:"required"`
	ProposedAt uint64         `json:"proposedAt" gencodec:"required"`
}

type batchMetadataMarshaling struct {
	BatchId    math.HexOrDecimal64 `json:"batchId" gencodec:"required"`
	ProposedAt math.HexOrDecimal64 `json:"proposedAt" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type LibSharedDataBaseFeeConfig -field-override libSharedDataBaseFeeConfigMarshaling -out gen_lib_shared_data_base_fee_config.go

// LibSharedDataBaseFeeConfig is an auto generated low-level Go binding around an user-defined struct.
type LibSharedDataBaseFeeConfig struct {
	AdjustmentQuotient     uint8  `json:"adjustmentQuotient" gencodec:"required"`
	SharingPctg            uint8  `json:"sharingPctg" gencodec:"required"`
	GasIssuancePerSecond   uint32 `json:"gasIssuancePerSecond" gencodec:"required"`
	MinGasExcess           uint64 `json:"minGasExcess" gencodec:"required"`
	MaxGasIssuancePerBlock uint32 `json:"maxGasIssuancePerBlock" gencodec:"required"`
}

type libSharedDataBaseFeeConfigMarshaling struct {
	MinGasExcess math.HexOrDecimal64 `json:"minGasExcess" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BlockParams  -out gen_block_params.go

// BlockParams is an auto generated low-level Go binding around an user-defined struct.
type BlockParams struct {
	NumTransactions uint16        `json:"numTransactions" gencodec:"required"`
	TimeShift       uint8         `json:"timeShift" gencodec:"required"`
	SignalSlots     []common.Hash `json:"signalSlots" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BatchInfo -field-override batchInfoMarshaling -out gen_batch_info.go

// BatchInfo is an auto generated low-level Go binding around an user-defined struct.
type BatchInfo struct {
	TxsHash            common.Hash
	Blocks             []*BlockParams
	BlobHashes         []common.Hash
	ExtraData          common.Hash
	Coinbase           common.Address
	ProposedIn         uint64
	BlobByteOffset     uint32
	BlobByteSize       uint32
	GasLimit           uint32
	LastBlockId        uint64
	LastBlockTimestamp uint64
	AnchorBlockId      uint64
	AnchorBlockHash    common.Hash
	BaseFeeConfig      *LibSharedDataBaseFeeConfig
}

type batchInfoMarshaling struct {
	ProposedIn         math.HexOrDecimal64 `json:"proposedIn" gencodec:"required"`
	LastBlockId        math.HexOrDecimal64 `json:"lastBlockId" gencodec:"required"`
	LastBlockTimestamp math.HexOrDecimal64 `json:"lastBlockTimestamp" gencodec:"required"`
	AnchorBlockId      math.HexOrDecimal64 `json:"anchorBlockId" gencodec:"required"`
}
