package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

//go:generate go run github.com/fjl/gencodec -type BatchProposed -field-override batchProposedMarshaling -out gen_batch_proposed.go

// BatchProposed represents a BatchProposed event raised by the TaikoInboxClient contract.
type BatchProposed struct {
	Info   *BatchInfo     `json:"info"   gencodec:"required"`
	Meta   *BatchMetadata `json:"meta"   gencodec:"required"`
	TxList hexutil.Bytes  `json:"txList" gencodec:"required"`
}

func (b *BatchProposed) GethType() *pacaya.TaikoInboxClientBatchProposed {
	if b == nil {
		log.Warn("missing BatchProposed when converting to GethType")
		return nil
	}
	blocks := make([]pacaya.ITaikoInboxBlockParams, len(b.Info.Blocks))
	for i, block := range b.Info.Blocks {
		blocks[i] = pacaya.ITaikoInboxBlockParams(*block)
	}

	return &pacaya.TaikoInboxClientBatchProposed{
		Info: pacaya.ITaikoInboxBatchInfo{
			TxsHash:            b.Info.TxsHash,
			Blocks:             blocks,
			BlobHashes:         b.Info.BlobHashes,
			ExtraData:          b.Info.ExtraData,
			Coinbase:           b.Info.Coinbase,
			ProposedIn:         b.Info.ProposedIn,
			BlobByteOffset:     b.Info.BlobByteOffset,
			BlobByteSize:       b.Info.BlobByteSize,
			GasLimit:           b.Info.GasLimit,
			LastBlockId:        b.Info.LastBlockID,
			LastBlockTimestamp: b.Info.LastBlockTimestamp,
			AnchorBlockId:      b.Info.AnchorBlockID,
			AnchorBlockHash:    b.Info.AnchorBlockHash,
			BaseFeeConfig:      pacaya.LibSharedDataBaseFeeConfig(*b.Info.BaseFeeConfig),
			BlobCreatedIn:      b.Info.BlobCreatedIn,
		},
		Meta: pacaya.ITaikoInboxBatchMetadata{
			InfoHash:   b.Meta.InfoHash,
			Proposer:   b.Meta.Proposer,
			BatchId:    b.Meta.BatchID,
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
	InfoHash   [32]byte       `json:"infoHash"   gencodec:"required"`
	Proposer   common.Address `json:"proposer"   gencodec:"required"`
	BatchID    uint64         `json:"batchId"    gencodec:"required"`
	ProposedAt uint64         `json:"proposedAt" gencodec:"required"`
}

type batchMetadataMarshaling struct {
	InfoHash   common.Hash         `json:"infoHash"   gencodec:"required"`
	BatchID    math.HexOrDecimal64 `json:"batchId"    gencodec:"required"`
	ProposedAt math.HexOrDecimal64 `json:"proposedAt" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type LibSharedDataBaseFeeConfig -field-override libSharedDataBaseFeeConfigMarshaling -out gen_lib_shared_data_base_fee_config.go

// LibSharedDataBaseFeeConfig is an auto generated low-level Go binding around an user-defined struct.
type LibSharedDataBaseFeeConfig struct {
	AdjustmentQuotient     uint8  `json:"adjustmentQuotient"     gencodec:"required"`
	SharingPctg            uint8  `json:"sharingPctg"            gencodec:"required"`
	GasIssuancePerSecond   uint32 `json:"gasIssuancePerSecond"   gencodec:"required"`
	MinGasExcess           uint64 `json:"minGasExcess"           gencodec:"required"`
	MaxGasIssuancePerBlock uint32 `json:"maxGasIssuancePerBlock" gencodec:"required"`
}

type libSharedDataBaseFeeConfigMarshaling struct {
	MinGasExcess math.HexOrDecimal64 `json:"minGasExcess" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BlockParams  -field-override  blockParamsMarshaling -out gen_block_params.go

// BlockParams is an auto generated low-level Go binding around an user-defined struct.
type BlockParams struct {
	NumTransactions uint16     `json:"numTransactions" gencodec:"required"`
	TimeShift       uint8      `json:"timeShift"       gencodec:"required"`
	SignalSlots     [][32]byte `json:"signalSlots"     gencodec:"required"`
}

type blockParamsMarshaling struct {
	SignalSlots []common.Hash `json:"signalSlots" gencodec:"required"`
}

//go:generate go run github.com/fjl/gencodec -type BatchInfo -field-override batchInfoMarshaling -out gen_batch_info.go

// BatchInfo is an auto generated low-level Go binding around an user-defined struct.
type BatchInfo struct {
	TxsHash            common.Hash                 `json:"txsHash"            gencodec:"required"`
	Blocks             []*BlockParams              `json:"blocks"             gencodec:"required"`
	BlobHashes         [][32]byte                  `json:"blobHashes"         gencodec:"required"`
	ExtraData          common.Hash                 `json:"extraData"          gencodec:"required"`
	Coinbase           common.Address              `json:"coinbase"           gencodec:"required"`
	ProposedIn         uint64                      `json:"proposedIn"         gencodec:"required"`
	BlobByteOffset     uint32                      `json:"blobByteOffset"     gencodec:"required"`
	BlobByteSize       uint32                      `json:"blobByteSize"       gencodec:"required"`
	BlobCreatedIn      uint64                      `json:"blobCreatedIn"      gencodec:"required"`
	GasLimit           uint32                      `json:"gasLimit"           gencodec:"required"`
	LastBlockID        uint64                      `json:"lastBlockId"        gencodec:"required"`
	LastBlockTimestamp uint64                      `json:"lastBlockTimestamp" gencodec:"required"`
	AnchorBlockID      uint64                      `json:"anchorBlockId"      gencodec:"required"`
	AnchorBlockHash    common.Hash                 `json:"anchorBlockHash"    gencodec:"required"`
	BaseFeeConfig      *LibSharedDataBaseFeeConfig `json:"baseFeeConfig"      gencodec:"required"`
}

type batchInfoMarshaling struct {
	BlobHashes         []common.Hash       `json:"blobHashes"         gencodec:"required"`
	ProposedIn         math.HexOrDecimal64 `json:"proposedIn"         gencodec:"required"`
	BlobCreatedIn      math.HexOrDecimal64 `json:"blobCreatedIn"      gencodec:"required"`
	LastBlockID        math.HexOrDecimal64 `json:"lastBlockId"        gencodec:"required"`
	LastBlockTimestamp math.HexOrDecimal64 `json:"lastBlockTimestamp" gencodec:"required"`
	AnchorBlockID      math.HexOrDecimal64 `json:"anchorBlockId"      gencodec:"required"`
}
