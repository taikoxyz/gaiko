package transition

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

const (
	HeklaHardFork  string = "Hekla"
	OntakeHardFork string = "Ontake"
	PacayaHardFork string = "Pacaya"
)

type BlockProposedFork interface {
	ABIEncoder
	BlockNumber() uint64
	BlockTimestamp() uint64
	BaseFeeConfig() pacaya.LibSharedDataBaseFeeConfig
	BlobTxSliceParam() (offset uint32, length uint32)
	BlobHash() common.Hash
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
	BatchInfo() *pacaya.ITaikoInboxBatchInfo
	GasLimit() uint32
	Coinbase() common.Address
	BlobHashes() [][32]byte
	ExtraData() [32]byte
}

func convertBaseFeeConfig(baseFeeConfig pacaya.LibSharedDataBaseFeeConfig) ontake.LibSharedDataBaseFeeConfig {
	return ontake.LibSharedDataBaseFeeConfig{
		AdjustmentQuotient:     baseFeeConfig.AdjustmentQuotient,
		SharingPctg:            baseFeeConfig.SharingPctg,
		GasIssuancePerSecond:   baseFeeConfig.GasIssuancePerSecond,
		MinGasExcess:           baseFeeConfig.MinGasExcess,
		MaxGasIssuancePerBlock: baseFeeConfig.MaxGasIssuancePerBlock,
	}
}

type PacayaBlockProposed struct {
	*pacaya.TaikoInboxClientBatchProposed
}

func (b *PacayaBlockProposed) Encode() ([]byte, error) {
	return batchProposedEvent.Inputs.Pack(b.Info, b.Meta, b.TxList)
}

func (b *PacayaBlockProposed) BlockNumber() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlockTimestamp() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BaseFeeConfig() pacaya.LibSharedDataBaseFeeConfig {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobTxSliceParam() (offset uint32, length uint32) {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobHash() common.Hash {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobUsed() bool {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) HardFork() HardFork {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) MinTier() uint16 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) ParentMetaHash() [32]byte {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) Sender() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) Difficulty() [32]byte {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) Proposer() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) LivenessBond() *big.Int {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) ProposedAt() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) ProposedIn() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobTxListOffset() uint32 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobTxListLength() uint32 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobIndex() uint8 {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) GasLimit() uint32 {
	panic("not implemented") // TODO: Implement
}
func (b *PacayaBlockProposed) Coinbase() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *PacayaBlockProposed) BlobHashes() [][32]byte {
	panic("not implemented")
}

func (b *PacayaBlockProposed) ExtraData() [32]byte {
	panic("not implemented")
}

type HeklaBlockProposed struct {
	*ontake.TaikoL1ClientBlockProposed
}

func (b *HeklaBlockProposed) BlockNumber() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlockTimestamp() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BaseFeeConfig() pacaya.LibSharedDataBaseFeeConfig {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobTxSliceParam() (offset uint32, length uint32) {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobHash() common.Hash {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobUsed() bool {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) HardFork() HardFork {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) MinTier() uint16 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) ParentMetaHash() [32]byte {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) Sender() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) Difficulty() [32]byte {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) Proposer() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) LivenessBond() *big.Int {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) ProposedAt() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) ProposedIn() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobTxListOffset() uint32 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobTxListLength() uint32 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobIndex() uint8 {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) GasLimit() uint32 {
	panic("not implemented") // TODO: Implement
}
func (b *HeklaBlockProposed) Coinbase() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *HeklaBlockProposed) BlobHashes() [][32]byte {
	panic("not implemented")
}

func (b *HeklaBlockProposed) ExtraData() [32]byte {
	panic("not implemented")
}

type OntakeBlockProposed struct {
	*ontake.TaikoL1ClientBlockProposedV2
}

func (b *OntakeBlockProposed) BlockNumber() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlockTimestamp() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BaseFeeConfig() pacaya.LibSharedDataBaseFeeConfig {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobTxSliceParam() (offset uint32, length uint32) {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobHash() common.Hash {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobUsed() bool {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) HardFork() HardFork {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) MinTier() uint16 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) ParentMetaHash() [32]byte {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) Sender() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) Difficulty() [32]byte {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) Proposer() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) LivenessBond() *big.Int {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) ProposedAt() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) ProposedIn() uint64 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobTxListOffset() uint32 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobTxListLength() uint32 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobIndex() uint8 {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) GasLimit() uint32 {
	panic("not implemented") // TODO: Implement
}
func (b *OntakeBlockProposed) Coinbase() common.Address {
	panic("not implemented") // TODO: Implement
}

func (b *OntakeBlockProposed) BlobHashes() [][32]byte {
	panic("not implemented")
}

func (b *OntakeBlockProposed) ExtraData() [32]byte {
	panic("not implemented")
}
