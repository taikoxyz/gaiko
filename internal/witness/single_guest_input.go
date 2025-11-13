package witness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"math"
	"math/big"
	"slices"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/gaiko/pkg/mpt"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

var (
	_ GuestInput       = (*SingleGuestInput)(nil)
	_ json.Unmarshaler = (*SingleGuestInput)(nil)
)

type SingleGuestInput struct {
	parent          *BatchGuestInput
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
	BlockProposed  BlockProposed
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
	ActualProver         common.Address            `json:"actual_prover"`
	Graffiti             common.Hash               `json:"graffiti"`
	ParentTransitionHash *common.Hash              `json:"parent_transition_hash"`
	Checkpoint           *ShastaProposalCheckpoint `json:"checkpoint"`

	DesignatedProver    common.Address `json:"-"`
	designatedProverSet bool           `json:"-"`
}

func (d *TaikoProverData) UnmarshalJSON(data []byte) error {
	type alias struct {
		ActualProver         common.Address            `json:"actual_prover"`
		Graffiti             common.Hash               `json:"graffiti"`
		ParentTransitionHash *common.Hash              `json:"parent_transition_hash"`
		Checkpoint           *ShastaProposalCheckpoint `json:"checkpoint"`
	}

	var base alias
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}

	d.ActualProver = base.ActualProver
	d.Graffiti = base.Graffiti
	d.ParentTransitionHash = base.ParentTransitionHash
	d.Checkpoint = base.Checkpoint

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if d.ActualProver == (common.Address{}) {
		if msg, ok := raw["prover"]; ok {
			trimmed := bytes.TrimSpace(msg)
			if len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null")) {
				var fallback common.Address
				if err := json.Unmarshal(trimmed, &fallback); err != nil {
					return err
				}
				d.ActualProver = fallback
			}
		}
	}

	if msg, ok := raw["designated_prover"]; ok {
		trimmed := bytes.TrimSpace(msg)
		if bytes.Equal(trimmed, []byte("null")) || len(trimmed) == 0 {
			d.DesignatedProver = common.Address{}
			d.designatedProverSet = false
		} else {
			var addr common.Address
			if err := json.Unmarshal(trimmed, &addr); err != nil {
				return err
			}
			d.DesignatedProver = addr
			d.designatedProverSet = true
		}
	} else {
		d.DesignatedProver = common.Address{}
		d.designatedProverSet = false
	}

	return nil
}

func (g *SingleGuestInput) GuestInputs() iter.Seq[*Pair] {
	return func(yield func(*Pair) bool) {
		blockProposed := g.Taiko.BlockProposed
		blobUsed := blockProposed.BlobUsed()
		compressedTxListBuf := g.Taiko.TxData
		var txs types.Transactions
		if g.IsTaiko() {
			if blobUsed {
				blob := eth.Blob(compressedTxListBuf)
				if blobDataBuf, err := blob.ToData(); err != nil {
					log.Error("Parse blob data failed", "err", err)
				} else if len(blobDataBuf) > maxBlobDataSize {
					panic(fmt.Sprintf("Blob data size exceeds the limit: %d", len(blobDataBuf)))
				} else if compressedTxListBuf, err = sliceTxList(g.Block.Number(), blobDataBuf, blockProposed.BlobTxSliceParam()); err != nil {
					log.Warn(
						"Invalid txlist offset and size in metadata",
						"blockID", g.Block.NumberU64(),
						"err", err,
					)
				}
				txs = decompressTxList(
					compressedTxListBuf,
					blobMaxTxListBytes,
					blobUsed,
				)
			} else {
				txs = decompressTxList(
					compressedTxListBuf,
					calldataMaxTxListBytes,
					blobUsed,
				)
			}
		} else {
			txs = decompressTxList(
				compressedTxListBuf,
				math.MaxUint64,
				blobUsed,
			)
		}
		txs = slices.Insert(txs, 0, g.Taiko.AnchorTx)
		if !yield(&Pair{g, txs}) {
			return
		}
	}
}

func (g *SingleGuestInput) BlockProposed() BlockProposed {
	return g.Taiko.BlockProposed
}

func (g *SingleGuestInput) Verify(proofType ProofType) error {
	if err := defaultSupportedChainSpecs.verifyChainSpec(g.ChainSpec); err != nil {
		return err
	}
	if g.Taiko.BlockProposed.BlobUsed() {
		if len(g.Taiko.TxData) != eth.BlobSize {
			return fmt.Errorf(
				"invalid TxData length, expected: %d, got: %d",
				eth.BlobSize, len(g.Taiko.TxData),
			)
		}
		blobProofType := getBlobProofType(proofType, g.Taiko.BlobProofType)
		var blob eth.Blob
		copy(blob[:], g.Taiko.TxData)
		if err := verifyBlob(blobProofType, &blob, *g.Taiko.BlobCommitment, (*kzg4844.Proof)(g.Taiko.BlobProof)); err != nil {
			return err
		}
	}
	return nil
}

func (g *SingleGuestInput) BlockMetadata() (BlockMetadata, error) {
	var (
		reducedGasLimit uint32
		txListHash      common.Hash
		metadata        BlockMetadata
	)
	if g.IsTaiko() {
		reducedGasLimit = anchorGasLimit
	}

	if g.Taiko.BlockProposed.BlobUsed() {
		if g.Taiko.BlobCommitment == nil {
			return nil, errors.New("missing blob commitment")
		}
		commitment := kzg4844.Commitment(*g.Taiko.BlobCommitment)
		txListHash = eth.KZGToVersionedHash(commitment)
	} else {
		txListHash = keccak.Keccak(g.Taiko.TxData)
	}

	var extraData [32]byte
	copy(extraData[:], g.Block.Extra())
	switch g.Taiko.BlockProposed.HardFork() {
	case NothingHardFork:
		metadata = &NothingBlockMetadata{}
	case HeklaHardFork:
		metadata = NewHeklaBlockMetadata(&ontake.TaikoDataBlockMetadata{
			L1Hash:         g.Taiko.L1Header.Hash(),
			Difficulty:     g.Taiko.BlockProposed.Difficulty(),
			BlobHash:       txListHash,
			ExtraData:      extraData,
			DepositsHash:   emptyEthDepositHash,
			Coinbase:       g.Block.Coinbase(),
			Id:             g.Block.NumberU64(),
			GasLimit:       uint32(g.Block.GasLimit()) - reducedGasLimit,
			Timestamp:      g.Block.Time(),
			L1Height:       g.Taiko.L1Header.Number.Uint64(),
			MinTier:        g.Taiko.BlockProposed.MinTier(),
			BlobUsed:       g.Taiko.BlockProposed.BlobUsed(),
			ParentMetaHash: g.Taiko.BlockProposed.ParentMetaHash(),
			Sender:         g.Taiko.BlockProposed.Sender(),
		})
	case OntakeHardFork:
		metadata = NewOntakeBlockMetadata(&ontake.TaikoDataBlockMetadataV2{
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
			BaseFeeConfig: ontake.LibSharedDataBaseFeeConfig(
				*g.Taiko.BlockProposed.BaseFeeConfig(),
			),
		})
	default:
		return nil, fmt.Errorf("unsupported hardfork: %s", g.Taiko.BlockProposed.HardFork())
	}
	return metadata, nil
}

func (g *SingleGuestInput) Transition() any {
	return &ontake.TaikoDataTransition{
		ParentHash: g.ParentHeader.Hash(),
		BlockHash:  g.Block.Hash(),
		StateRoot:  g.Block.Root(),
		Graffiti:   g.Taiko.ProverData.Graffiti,
	}
}

func (g *SingleGuestInput) ForkVerifierAddress(proofType ProofType) common.Address {
	return g.ChainSpec.getForkVerifierAddress(
		g.Taiko.BlockProposed.BlockNumber(),
		g.Block.Time(),
		proofType,
	)
}

func (g *SingleGuestInput) Prover() common.Address {
	return g.Taiko.ProverData.ActualProver
}

func (g *SingleGuestInput) ChainID() uint64 {
	return g.ChainSpec.ChainID
}

func (g *SingleGuestInput) ID() ID {
	id := ID{
		BlockID: g.Block.NumberU64(),
	}
	if g.parent != nil {
		id.BatchID = g.parent.Taiko.BatchID
	}
	return id
}

func (g *SingleGuestInput) IsTaiko() bool {
	return g.ChainSpec.IsTaiko
}

func (g *SingleGuestInput) ChainConfig() (*params.ChainConfig, error) {
	return g.ChainSpec.chainConfig(false)
}
