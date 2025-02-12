package transition

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

type StorageEntry struct {
	Trie  trie.Trie
	Slots []*big.Int
}

type GuestInput struct {
	Block           types.Block                     `json:"block"`
	ChainSpec       ChainSpec                       `json:"chain_spec"`
	ParentHeader    types.Header                    `json:"parent_header"`
	ParentStateTrie trie.Trie                       `json:"parent_state_trie"`
	ParentStorage   map[common.Address]StorageEntry `json:"parent_storage"`
	Contracts       [][]byte                        `json:"contracts"`
	AncestorHeaders []types.Header                  `json:"ancestor_headers"`
	Taiko           TaikoGuestInput                 `json:"taiko"`
}

type TaikoGuestInput struct {
	L1Header       types.Header       `json:"l1_header"`
	TxData         []byte             `json:"tx_data"`
	AnchorTx       *types.Transaction `json:"anchor_tx"`
	BlockProposed  BlockProposedFork  `json:"block_proposed"`
	ProverData     TaikoProverData    `json:"prover_data"`
	BlobCommitment *[]byte            `json:"blob_commitment"`
	BlobProof      *[]byte            `json:"blob_proof"`
	BlobProofType  BlobProofType      `json:"blob_proof_type"`
}

type BlobProofType string

const (
	KzgVersionedHash   BlobProofType = "kzg_versioned_hash"
	ProofOfEquivalence BlobProofType = "proof_of_equivalence"
)

type TaikoProverData struct {
	Prover   common.Address
	Graffiti common.Hash
}

type TxSlicePosition struct {
	Offset uint
	Length uint
}

type hardFork string

const (
	HeklaHardFork  hardFork = "Hekla"
	OntakeHardFork hardFork = "Ontake"
	PacayaHardFork hardFork = "Pacaya"
)

type BlockProposedFork interface {
	BlockNumber() uint64
	BlockTimestamp() uint64
	BaseFeeConfig() ontake.LibSharedDataBaseFeeConfig
	BlobTxSliceParam() *TxSlicePosition
	BlobHash() common.Hash
	BlobUsed() bool
	HardFork() hardFork
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
}

//go:generate go run github.com/fjl/gencodec -type TaikoL1BlockProposed -field-override taikoL1BlockProposedMarshaling -out gen_taiko_l1_block_proposed.go
type TaikoL1BlockProposed struct {
	BlockId           *big.Int
	AssignedProver    common.Address
	LivenessBond      *big.Int
	Meta              TaikoDataBlockMetadata
	DepositsProcessed []TaikoDataEthDeposit
	Raw               types.Log // Blockchain specific contextual infos
}

type taikoL1BlockProposedMarshaling struct {
	BlockId      *hexutil.Big
	LivenessBond *hexutil.Big
}

func (b *TaikoL1BlockProposed) BlobUsed() bool {
	return b.Meta.BlobUsed
}

func (b *TaikoL1BlockProposed) BlockNumber() uint64 {
	return b.Meta.Id
}

func (b *TaikoL1BlockProposed) BlockTimestamp() uint64 {
	return b.Meta.Timestamp
}

func (b *TaikoL1BlockProposed) BaseFeeConfig() ontake.LibSharedDataBaseFeeConfig {
	return ontake.LibSharedDataBaseFeeConfig{}
}

func (b *TaikoL1BlockProposed) BlobTxSliceParam() *TxSlicePosition {
	return nil
}

func (b *TaikoL1BlockProposed) BlobHash() common.Hash {
	return b.Meta.BlobHash
}

func (b *TaikoL1BlockProposed) IsPacaya() bool {
	return false
}

type TaikoL1BlockProposedV2 struct {
	BlockId *big.Int
	Meta    TaikoDataBlockMetadataV2
	Raw     types.Log // Blockchain specific contextual infos
}

func (b *TaikoL1BlockProposedV2) BlobUsed() bool {
	return b.Meta.BlobUsed
}

func (b *TaikoL1BlockProposedV2) BlockNumber() uint64 {
	return b.Meta.Id
}

func (b *TaikoL1BlockProposedV2) BlockTimestamp() uint64 {
	return b.Meta.Timestamp
}

func (b *TaikoL1BlockProposedV2) BaseFeeConfig() ontake.LibSharedDataBaseFeeConfig {
	return b.Meta.BaseFeeConfig
}

func (b *TaikoL1BlockProposedV2) BlobTxSliceParam() *TxSlicePosition {
	return &TxSlicePosition{
		Offset: uint(b.Meta.BlobTxListOffset),
		Length: uint(b.Meta.BlobTxListLength),
	}
}

func (b *TaikoL1BlockProposedV2) BlobHash() common.Hash {
	return b.Meta.BlobHash
}

func (b *TaikoL1BlockProposedV2) IsPacaya() bool {
	return false
}

type TaikoDataBlockMetadataV2 struct {
	AnchorBlockHash  [32]byte
	Difficulty       [32]byte
	BlobHash         [32]byte
	ExtraData        [32]byte
	Coinbase         common.Address
	Id               uint64
	GasLimit         uint32
	Timestamp        uint64
	AnchorBlockId    uint64
	MinTier          uint16
	BlobUsed         bool
	ParentMetaHash   [32]byte
	Proposer         common.Address
	LivenessBond     *big.Int
	ProposedAt       uint64
	ProposedIn       uint64
	BlobTxListOffset uint32
	BlobTxListLength uint32
	BlobIndex        uint8
	BaseFeeConfig    ontake.LibSharedDataBaseFeeConfig
}

type TaikoDataEthDeposit struct {
	Recipient common.Address
	Amount    *big.Int
	Id        uint64
}

//go:generate go run github.com/fjl/gencodec -type TaikoDataBlockMetadata -field-override taikoDataBlockMetadataMarshaling -out gen_taiko_data_block_metadata.go
type TaikoDataBlockMetadata struct {
	L1Hash         [32]byte
	Difficulty     [32]byte
	BlobHash       [32]byte
	ExtraData      [32]byte
	DepositsHash   [32]byte
	Coinbase       common.Address
	Id             uint64
	GasLimit       uint32
	Timestamp      uint64
	L1Height       uint64
	MinTier        uint16
	BlobUsed       bool
	ParentMetaHash [32]byte
	Sender         common.Address
}

type taikoDataBlockMetadataMarshaling struct {
	BlockId      *hexutil.Big
	LivenessBond *hexutil.Big
}

type SpecId = uint8
type ProofType = uint8

const (
	SgxProofType ProofType = 2
)

type ChainSpec struct {
	Name                 string
	ChainId              uint64
	MaxSpecId            SpecId
	HardForks            map[SpecId]ForkCondition
	Eip1559Constants     Eip1559Constants
	L1Contract           *common.Address
	L2Contract           *common.Address
	RPC                  string
	BeaconRPC            *string
	VerifierAddressForks map[SpecId]map[ProofType]*common.Address
	GenesisTime          uint64
	SecondsPerSlot       uint64
	IsTaiko              bool
}

// pub fn get_fork_verifier_address(
// 	&self,
// 	block_num: u64,
// 	proof_type: ProofType,
// ) -> Result<Address> {
// 	// fall down to the first fork that is active as default
// 	for (spec_id, fork) in self.hard_forks.iter().rev() {
// 		if fork.active(block_num, 0u64) {
// 			if let Some(fork_verifier) = self.verifier_address_forks.get(spec_id) {
// 				return fork_verifier
// 					.get(&proof_type)
// 					.ok_or_else(|| anyhow!("Verifier type not found"))
// 					.and_then(|address| {
// 						address.ok_or_else(|| anyhow!("Verifier address not found"))
// 					});
// 			}
// 		}
// 	}

// 	Err(anyhow!("fork verifier is not active"))
// }

func (c *ChainSpec) getForkVerifierAddress(blockNum uint64) (common.Address, error) {
	// TODO: Implement this function
	return common.Address{}, nil
}

type ForkCondition interface {
	Active(blockNumber uint64, timestamp uint64) bool
}

type BlockNumber uint64

func (b BlockNumber) Active(blockNumber uint64, _ uint64) bool {
	return blockNumber >= uint64(b)
}

type BlockTimestamp uint64

func (b BlockTimestamp) Active(_ uint64, timestamp uint64) bool {
	return timestamp >= uint64(b)
}

type Eip1559Constants struct {
	BaseFeeChangeDenominator      *big.Int
	BaseFeeMaxIncreaseDenominator *big.Int
	BaseFeeMaxDecreaseDenominator *big.Int
	ElasticityMultiplier          *big.Int
}
