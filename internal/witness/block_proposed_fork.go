package witness

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

const (
	NothingHardFork string = "Nothing"
	HeklaHardFork   string = "Hekla"
	OntakeHardFork  string = "Ontake"
	PacayaHardFork  string = "Pacaya"
	ShastaHardFork  string = "Shasta"
)

// Slice represents the offset and length of a slice.
type Slice struct {
	Offset uint32
	Length uint32
}

// BlockProposedFork represents the interface for handling taiko proposed blocks in different hard forks.
type BlockProposedFork interface {
	ABIEncoder
	BlockNumber() uint64
	BlockTimestamp() uint64
	BaseFeeConfig() *pacaya.LibSharedDataBaseFeeConfig
	BlobTxSliceParam() *Slice
	BlobUsed() bool
	HardFork() string
	MinTier() uint16
	ParentMetaHash() [32]byte
	Sender() common.Address
	Difficulty() [32]byte
	Proposer() common.Address
	LivenessBond() *big.Int
	ProposedAt() uint64
	ProposedIn() uint64
	BlobTxListOffset() uint32
	BlobTxListLength() uint32
	BlobIndex() uint8
	GasLimit() uint32
	Coinbase() common.Address
	BlobHashes() [][32]byte
	ExtraData() [32]byte
	BlockParams() []*pacaya.ITaikoInboxBlockParams
	BlobCreatedIn() uint64
	BlockMetadataFork() BlockMetadataFork
}

var _ BlockProposedFork = (*PacayaBlockProposed)(nil)

type PacayaBlockProposed struct {
	*pacaya.TaikoInboxClientBatchProposed
	blockParams []*pacaya.ITaikoInboxBlockParams
}

func NewPacayaBlockProposed(b *pacaya.TaikoInboxClientBatchProposed) *PacayaBlockProposed {
	blockParams := make([]*pacaya.ITaikoInboxBlockParams, len(b.Info.Blocks))
	for i, block := range b.Info.Blocks {
		blockParams[i] = &block
	}

	return &PacayaBlockProposed{
		TaikoInboxClientBatchProposed: b,
		blockParams:                   blockParams,
	}
}

func (b *PacayaBlockProposed) ABIEncode() ([]byte, error) {
	return batchProposedEvent.Inputs.Pack(b.Info, b.Meta, b.TxList)
}

func (b *PacayaBlockProposed) BlockNumber() uint64 {
	return b.Info.LastBlockId - uint64(len(b.Info.Blocks)) + 1
}

func (b *PacayaBlockProposed) BlockTimestamp() uint64 {
	return 0
}

func (b *PacayaBlockProposed) BaseFeeConfig() *pacaya.LibSharedDataBaseFeeConfig {
	return &b.Info.BaseFeeConfig
}

func (b *PacayaBlockProposed) BlobTxSliceParam() *Slice {
	return &Slice{b.Info.BlobByteOffset, b.Info.BlobByteSize}
}

func (b *PacayaBlockProposed) BlobUsed() bool {
	return len(b.BlobHashes()) > 0
}

func (b *PacayaBlockProposed) HardFork() string {
	return PacayaHardFork
}

func (b *PacayaBlockProposed) MinTier() uint16 {
	return 0
}

func (b *PacayaBlockProposed) ParentMetaHash() [32]byte {
	return [32]byte{}
}

func (b *PacayaBlockProposed) Sender() common.Address {
	return common.Address{}
}

func (b *PacayaBlockProposed) Difficulty() [32]byte {
	return [32]byte{}
}

func (b *PacayaBlockProposed) Proposer() common.Address {
	return b.Meta.Proposer
}

func (b *PacayaBlockProposed) LivenessBond() *big.Int {
	return nil
}

func (b *PacayaBlockProposed) ProposedAt() uint64 {
	return b.Meta.ProposedAt
}

func (b *PacayaBlockProposed) ProposedIn() uint64 {
	return b.Info.ProposedIn
}

func (b *PacayaBlockProposed) BlobTxListOffset() uint32 {
	return b.Info.BlobByteOffset
}

func (b *PacayaBlockProposed) BlobTxListLength() uint32 {
	return b.Info.BlobByteSize
}

func (b *PacayaBlockProposed) BlobIndex() uint8 {
	return 0
}

func (b *PacayaBlockProposed) GasLimit() uint32 {
	return b.Info.GasLimit
}
func (b *PacayaBlockProposed) Coinbase() common.Address {
	return b.Info.Coinbase
}

func (b *PacayaBlockProposed) BlobHashes() [][32]byte {
	return b.Info.BlobHashes
}

func (b *PacayaBlockProposed) ExtraData() [32]byte {
	return b.Info.ExtraData
}

func (b *PacayaBlockProposed) BlockParams() []*pacaya.ITaikoInboxBlockParams {
	return b.blockParams
}

func (b *PacayaBlockProposed) BlobCreatedIn() uint64 {
	return b.Info.BlobCreatedIn
}

func (b *PacayaBlockProposed) BlockMetadataFork() BlockMetadataFork {
	return NewPacayaBlockMetadata(&b.Meta)
}

type HeklaBlockProposed struct {
	*ontake.TaikoL1ClientBlockProposed
}

var _ BlockProposedFork = (*HeklaBlockProposed)(nil)

func NewHeklaBlockProposed(b *ontake.TaikoL1ClientBlockProposed) *HeklaBlockProposed {
	return &HeklaBlockProposed{b}
}

func (b *HeklaBlockProposed) ABIEncode() ([]byte, error) {
	return blockMetadataComponentsArgs.Pack(
		b.BlockId,
		b.AssignedProver,
		b.TaikoL1ClientBlockProposed.LivenessBond,
		b.Meta,
		b.DepositsProcessed,
	)
}

func (b *HeklaBlockProposed) BlockNumber() uint64 {
	return b.Meta.Id
}

func (b *HeklaBlockProposed) BlockTimestamp() uint64 {
	return b.Meta.Timestamp
}

func (b *HeklaBlockProposed) BaseFeeConfig() *pacaya.LibSharedDataBaseFeeConfig {
	return nil
}

func (b *HeklaBlockProposed) BlobTxSliceParam() *Slice {
	return nil
}

func (b *HeklaBlockProposed) BlobUsed() bool {
	return b.Meta.BlobUsed
}

func (b *HeklaBlockProposed) HardFork() string {
	return HeklaHardFork
}

func (b *HeklaBlockProposed) MinTier() uint16 {
	return b.Meta.MinTier
}

func (b *HeklaBlockProposed) ParentMetaHash() [32]byte {
	return b.Meta.ParentMetaHash
}

func (b *HeklaBlockProposed) Sender() common.Address {
	return b.Meta.Sender
}

func (b *HeklaBlockProposed) Difficulty() [32]byte {
	return b.Meta.Difficulty
}

func (b *HeklaBlockProposed) Proposer() common.Address {
	return common.Address{}
}

func (b *HeklaBlockProposed) LivenessBond() *big.Int {
	return b.TaikoL1ClientBlockProposed.LivenessBond
}

func (b *HeklaBlockProposed) ProposedAt() uint64 {
	return 0
}

func (b *HeklaBlockProposed) ProposedIn() uint64 {
	return 0
}

func (b *HeklaBlockProposed) BlobTxListOffset() uint32 {
	return 0
}

func (b *HeklaBlockProposed) BlobTxListLength() uint32 {
	return 0
}

func (b *HeklaBlockProposed) BlobIndex() uint8 {
	return 0
}

func (b *HeklaBlockProposed) GasLimit() uint32 {
	return b.Meta.GasLimit
}

func (b *HeklaBlockProposed) Coinbase() common.Address {
	return b.Meta.Coinbase
}

func (b *HeklaBlockProposed) BlobHashes() [][32]byte {
	return nil
}

func (b *HeklaBlockProposed) ExtraData() [32]byte {
	return b.Meta.ExtraData
}

func (b *HeklaBlockProposed) BlockParams() []*pacaya.ITaikoInboxBlockParams {
	return nil
}

func (b *HeklaBlockProposed) BlobCreatedIn() uint64 {
	return 0
}

func (b *HeklaBlockProposed) BlockMetadataFork() BlockMetadataFork {
	return NewHeklaBlockMetadata(&b.Meta)
}

type OntakeBlockProposed struct {
	*ontake.TaikoL1ClientBlockProposedV2
}

var _ BlockProposedFork = (*OntakeBlockProposed)(nil)

func NewOntakeBlockProposed(b *ontake.TaikoL1ClientBlockProposedV2) *OntakeBlockProposed {
	return &OntakeBlockProposed{b}
}

func (b *OntakeBlockProposed) ABIEncode() ([]byte, error) {
	return blockMetadataV2ComponentsArgs.Pack(b.BlockId, b.Meta)
}

func (b *OntakeBlockProposed) BlockNumber() uint64 {
	return b.Meta.Id
}

func (b *OntakeBlockProposed) BlockTimestamp() uint64 {
	return b.Meta.Timestamp
}

func (b *OntakeBlockProposed) BaseFeeConfig() *pacaya.LibSharedDataBaseFeeConfig {
	return (*pacaya.LibSharedDataBaseFeeConfig)(&b.Meta.BaseFeeConfig)
}

func (b *OntakeBlockProposed) BlobTxSliceParam() *Slice {
	return &Slice{b.Meta.BlobTxListOffset, b.Meta.BlobTxListLength}
}

func (b *OntakeBlockProposed) BlobUsed() bool {
	return b.Meta.BlobUsed
}

func (b *OntakeBlockProposed) HardFork() string {
	return OntakeHardFork
}

func (b *OntakeBlockProposed) MinTier() uint16 {
	return b.Meta.MinTier
}

func (b *OntakeBlockProposed) ParentMetaHash() [32]byte {
	return b.Meta.ParentMetaHash
}

func (b *OntakeBlockProposed) Sender() common.Address {
	return common.Address{}
}

func (b *OntakeBlockProposed) Difficulty() [32]byte {
	return b.Meta.Difficulty
}

func (b *OntakeBlockProposed) Proposer() common.Address {
	return b.Meta.Proposer
}

func (b *OntakeBlockProposed) LivenessBond() *big.Int {
	return b.Meta.LivenessBond
}

func (b *OntakeBlockProposed) ProposedAt() uint64 {
	return b.Meta.ProposedAt
}

func (b *OntakeBlockProposed) ProposedIn() uint64 {
	return b.Meta.ProposedIn
}

func (b *OntakeBlockProposed) BlobTxListOffset() uint32 {
	return b.Meta.BlobTxListOffset
}

func (b *OntakeBlockProposed) BlobTxListLength() uint32 {
	return b.Meta.BlobTxListLength
}

func (b *OntakeBlockProposed) BlobIndex() uint8 {
	return b.Meta.BlobIndex
}

func (b *OntakeBlockProposed) GasLimit() uint32 {
	return b.Meta.GasLimit
}

func (b *OntakeBlockProposed) Coinbase() common.Address {
	return b.Meta.Coinbase
}

func (b *OntakeBlockProposed) BlobHashes() [][32]byte {
	return nil
}

func (b *OntakeBlockProposed) ExtraData() [32]byte {
	return b.Meta.ExtraData
}

func (b *OntakeBlockProposed) BlockParams() []*pacaya.ITaikoInboxBlockParams {
	return nil
}

func (b *OntakeBlockProposed) BlobCreatedIn() uint64 {
	return 0
}

func (b *OntakeBlockProposed) BlockMetadataFork() BlockMetadataFork {
	return NewOntakeBlockMetadata(&b.Meta)
}

type NotingBlockProposed struct{}

var _ BlockProposedFork = (*NotingBlockProposed)(nil)

func (b *NotingBlockProposed) ABIEncode() ([]byte, error) {
	return nil, nil
}

func (b *NotingBlockProposed) BlockNumber() uint64 {
	return 0
}

func (b *NotingBlockProposed) BlockTimestamp() uint64 {
	return 0
}

func (b *NotingBlockProposed) BaseFeeConfig() *pacaya.LibSharedDataBaseFeeConfig {
	return nil
}

func (b *NotingBlockProposed) BlobTxSliceParam() *Slice {
	return nil
}

func (b *NotingBlockProposed) BlobUsed() bool {
	return false
}

func (b *NotingBlockProposed) HardFork() string {
	return NothingHardFork
}

func (b *NotingBlockProposed) MinTier() uint16 {
	return 0
}

func (b *NotingBlockProposed) ParentMetaHash() [32]byte {
	return [32]byte{}
}

func (b *NotingBlockProposed) Sender() common.Address {
	return common.Address{}
}

func (b *NotingBlockProposed) Difficulty() [32]byte {
	return [32]byte{}
}

func (b *NotingBlockProposed) Proposer() common.Address {
	return common.Address{}
}

func (b *NotingBlockProposed) LivenessBond() *big.Int {
	return nil
}

func (b *NotingBlockProposed) ProposedAt() uint64 {
	return 0
}

func (b *NotingBlockProposed) ProposedIn() uint64 {
	return 0
}

func (b *NotingBlockProposed) BlobTxListOffset() uint32 {
	return 0
}

func (b *NotingBlockProposed) BlobTxListLength() uint32 {
	return 0
}

func (b *NotingBlockProposed) BlobIndex() uint8 {
	return 0
}

func (b *NotingBlockProposed) GasLimit() uint32 {
	return 0
}

func (b *NotingBlockProposed) Coinbase() common.Address {
	return common.Address{}
}

func (b *NotingBlockProposed) BlobHashes() [][32]byte {
	return nil
}

func (b *NotingBlockProposed) ExtraData() [32]byte {
	return [32]byte{}
}

func (b *NotingBlockProposed) BlockParams() []*pacaya.ITaikoInboxBlockParams {
	return nil
}

func (b *NotingBlockProposed) BlobCreatedIn() uint64 {
	return 0
}

func (b *NotingBlockProposed) BlockMetadataFork() BlockMetadataFork {
	return nil
}
