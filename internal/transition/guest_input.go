package transition

import (
	"encoding/json"
	"fmt"
	"iter"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

var _ GuestDriver = (*GuestInput)(nil)
var _ json.Unmarshaler = (*GuestInput)(nil)

type GuestInput struct {
	Block           *types.Block
	ChainSpec       *ChainSpec
	ParentHeader    *types.Header
	ParentStateTrie *trie.Trie
	ParentStorage   map[common.Address]StorageEntry
	Contracts       [][]byte
	AncestorHeaders []*types.Header
	Taiko           *TaikoGuestInput
}

type StorageEntry struct {
	Trie  *trie.Trie
	Slots []*big.Int
}

type TaikoGuestInput struct {
	L1Header       *types.Header
	TxData         []byte
	AnchorTx       *types.Transaction
	BlockProposed  BlockProposedFork
	ProverData     *TaikoProverData
	BlobCommitment *[commitmentSize]byte
	BlobProof      *[proofSize]byte
	BlobProofType  BlobProofType
}

type BlobProofType string

const (
	KzgVersionedHash   BlobProofType = "kzg_versioned_hash"
	ProofOfEquivalence BlobProofType = "proof_of_equivalence"
)

type TaikoProverData struct {
	Prover   common.Address `json:"prover"`
	Graffiti common.Hash    `json:"graffiti"`
}

func (g *GuestInput) GuestInputs() iter.Seq[Pair] {
	return func(yield func(Pair) bool) {
		chainID := big.NewInt(int64(g.ChainSpec.ChainID))
		txListBytes := g.Taiko.TxData
		blobUsed := g.Taiko.BlockProposed.BlobUsed()
		isPacaya := g.Taiko.BlockProposed.HardFork() == PacayaHardFork
		offset, length := g.Taiko.BlockProposed.BlobTxSliceParam()
		txs := decodeTxs(txListBytes, g.Taiko.AnchorTx, blobUsed, isPacaya, chainID, g.Block.Number(), offset, length)
		yield(Pair{g, txs})
	}
}

func (g *GuestInput) BlockProposedFork() BlockProposedFork {
	return g.Taiko.BlockProposed
}

func (g *GuestInput) BlockMetaDataFork(proofType ProofType) (BlockMetaDataFork, error) {
	var (
		reducedGasLimit uint32
		txListHash      common.Hash
		metadata        BlockMetaDataFork
		blobProofType   = getBlobProofType(proofType, g.Taiko.BlobProofType)
	)
	if g.ChainSpec.IsTaiko {
		reducedGasLimit = anchorGasLimit
	}

	if g.Taiko.BlockProposed.BlobUsed() {
		if g.Taiko.BlobCommitment == nil {
			return nil, fmt.Errorf("missing blob commitment")
		}
		commitment := kzg4844.Commitment(*g.Taiko.BlobCommitment)
		txListHash = eth.KZGToVersionedHash(commitment)
		if len(g.Taiko.TxData) != eth.BlobSize {
			return nil, fmt.Errorf("invalid TxData length, expected: %d, got: %d", eth.BlobSize, len(g.Taiko.TxData))
		}
		var blob [eth.BlobSize]byte
		copy(blob[:], g.Taiko.TxData)
		if err := verifyBlob(blobProofType, blob, *g.Taiko.BlobCommitment, (*kzg4844.Proof)(g.Taiko.BlobProof)); err != nil {
			return nil, err
		}
	} else {
		txListHash = common.BytesToHash(keccak(g.Taiko.TxData))
	}

	var extraData [32]byte
	copy(extraData[:], g.Block.Extra())
	switch g.Taiko.BlockProposed.HardFork() {
	case HeklaHardFork:
		metadata = &HeklaBlockMetadata{
			TaikoDataBlockMetadata: &ontake.TaikoDataBlockMetadata{
				L1Hash:         g.Taiko.L1Header.Hash(),
				Difficulty:     g.Taiko.BlockProposed.Difficulty(),
				BlobHash:       txListHash,
				ExtraData:      extraData,
				DepositsHash:   emptyHash,
				Coinbase:       g.Block.Coinbase(),
				Id:             g.Block.NumberU64(),
				GasLimit:       uint32(g.Block.GasLimit()) - reducedGasLimit,
				Timestamp:      g.Block.Time(),
				L1Height:       g.Taiko.L1Header.Number.Uint64(),
				MinTier:        g.Taiko.BlockProposed.MinTier(),
				BlobUsed:       g.Taiko.BlockProposed.BlobUsed(),
				ParentMetaHash: g.Taiko.BlockProposed.ParentMetaHash(),
				Sender:         g.Taiko.BlockProposed.Sender(),
			},
		}
	case OntakeHardFork:
		metadata = &OntakeBlockMetadata{
			TaikoDataBlockMetadataV2: &ontake.TaikoDataBlockMetadataV2{
				AnchorBlockHash:  g.Taiko.L1Header.Hash(),
				Difficulty:       g.Taiko.BlockProposed.Difficulty(),
				BlobHash:         txListHash,
				ExtraData:        extraData,
				Coinbase:         g.Block.Coinbase(),
				Id:               g.Block.NumberU64(),
				GasLimit:         uint32(g.Block.GasLimit()) - reducedGasLimit,
				Timestamp:        g.Block.Time(),
				AnchorBlockId:    g.Taiko.L1Header.Number.Uint64(),
				MinTier:          g.Taiko.BlockProposed.MinTier(),
				BlobUsed:         g.Taiko.BlockProposed.BlobUsed(),
				ParentMetaHash:   g.Taiko.BlockProposed.ParentMetaHash(),
				Proposer:         g.Taiko.BlockProposed.Proposer(),
				LivenessBond:     g.Taiko.BlockProposed.LivenessBond(),
				ProposedAt:       g.Taiko.BlockProposed.ProposedAt(),
				ProposedIn:       g.Taiko.BlockProposed.ProposedIn(),
				BlobTxListOffset: g.Taiko.BlockProposed.BlobTxListOffset(),
				BlobTxListLength: g.Taiko.BlockProposed.BlobTxListLength(),
				BlobIndex:        g.Taiko.BlockProposed.BlobIndex(),
				BaseFeeConfig:    convertBaseFeeConfig(g.Taiko.BlockProposed.BaseFeeConfig()),
			},
		}
	default:
		return nil, fmt.Errorf("unsupported hardfork: %v", g.Taiko.BlockProposed.HardFork())
	}
	return metadata, nil
}

func (g *GuestInput) Transition() *ontake.TaikoDataTransition {
	return &ontake.TaikoDataTransition{
		ParentHash: g.ParentHeader.Hash(),
		BlockHash:  g.Block.Hash(),
		StateRoot:  g.Block.Root(),
		Graffiti:   g.Taiko.ProverData.Graffiti,
	}
}

func (g *GuestInput) ForkVerifierAddress(proofType ProofType) (common.Address, error) {
	return g.ChainSpec.getForkVerifierAddress(g.Taiko.BlockProposed.BlockNumber(), proofType)
}

func (g *GuestInput) Prover() common.Address {
	return g.Taiko.ProverData.Prover
}

func (g *GuestInput) ChainID() uint64 {
	return g.ChainSpec.ChainID
}

func (g *GuestInput) IsTaiko() bool {
	return g.ChainSpec.IsTaiko
}

func (g *GuestInput) ChainConfig() (*params.ChainConfig, error) {
	return g.ChainSpec.chainConfig()
}
