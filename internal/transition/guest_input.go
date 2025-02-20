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
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/gaiko/internal"
	"github.com/taikoxyz/gaiko/internal/mpt"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

var _ GuestDriver = (*GuestInput)(nil)
var _ json.Unmarshaler = (*GuestInput)(nil)

type GuestInput struct {
	Block           *types.Block
	ChainSpec       *ChainSpec
	ParentHeader    *types.Header
	ParentStateTrie *mpt.MptNode
	ParentStorage   map[common.Address]*StorageEntry
	Contracts       [][]byte
	AncestorHeaders []*types.Header
	Taiko           *TaikoGuestInput
}

type StorageEntry struct {
	Trie  *mpt.MptNode
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

func (g *GuestInput) GuestInputs() iter.Seq[*Pair] {
	return func(yield func(*Pair) bool) {
		chainID := big.NewInt(int64(g.ChainSpec.ChainID))
		txListBytes := g.Taiko.TxData
		blobUsed := g.Taiko.BlockProposed.BlobUsed()
		isPacaya := g.Taiko.BlockProposed.HardFork() == PacayaHardFork
		offset, length := g.Taiko.BlockProposed.BlobTxSliceParam()
		if blobUsed {
			blob := eth.Blob(txListBytes)
			var err error
			if txListBytes, err = blob.ToData(); err != nil {
				log.Warn("Parse blob data failed", "err", err)
				return
			}
			if txListBytes, err = sliceTxList(g.Block.Number(), txListBytes, offset, length); err != nil {
				log.Warn("Invalid txlist offset and size in metadata", "blockID", g.Block.NumberU64(), "err", err)
				return
			}
		}
		txs := decompressTxList(txListBytes, blobUsed, isPacaya, chainID)
		txs = append([]*types.Transaction{g.Taiko.AnchorTx}, txs...)
		if !yield(&Pair{g, txs}) {
			return
		}
	}
}

func (g *GuestInput) BlockProposedFork() BlockProposedFork {
	return g.Taiko.BlockProposed
}

func (g *GuestInput) BlockMetadataFork(proofType ProofType) (BlockMetadataFork, error) {
	var (
		reducedGasLimit uint32
		txListHash      common.Hash
		metadata        BlockMetadataFork
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
		var blob eth.Blob
		copy(blob[:], g.Taiko.TxData)
		if err := verifyBlob(blobProofType, &blob, *g.Taiko.BlobCommitment, (*kzg4844.Proof)(g.Taiko.BlobProof)); err != nil {
			return nil, err
		}
	} else {
		txListHash = common.BytesToHash(internal.Keccak(g.Taiko.TxData))
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
