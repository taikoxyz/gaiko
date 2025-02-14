package transition

import (
	"crypto/sha256"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

const (
	anchorGasLimit uint32 = 250000
)

var emptyHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

type BlockMetadataV2 struct {
	*ontake.TaikoDataBlockMetadataV2
}

func (m *BlockMetadataV2) Hash() common.Hash {
	b, _ := blockMetadataV2ComponentsArgs.Pack(m.TaikoDataBlockMetadataV2)
	return common.BytesToHash(keccak(b))
}

type publicInputs struct {
	transition     *ontake.TaikoDataTransition
	block_metadata BlockMetaDataFork
	verifier       common.Address
	prover         common.Address
	sgxInstance    common.Address
	chainID        uint64
}

func getBlobProofType(proofType ProofType, blobProofTypeHint BlobProofType) BlobProofType {
	switch proofType {
	case NativeProofType:
		return blobProofTypeHint
	case SgxProofType, GaikoSgxProofType:
		return KzgVersionedHash
	case Sp1ProofType, Risc0ProofType:
		return ProofOfEquivalence
	default:
		panic("unreachable")
	}
}

func (g *GuestInput) publicInputs(proofType ProofType) (*publicInputs, error) {
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
		var commitment kzg4844.Commitment
		copy(commitment[:], *g.Taiko.BlobCommitment)
		var blob kzg4844.Blob
		copy(blob[:], g.Taiko.TxData)

		txListHash = common.Hash(kzg4844.CalcBlobHashV1(sha256.New(), &commitment))
		switch blobProofType {
		case KzgVersionedHash:
			got, err := kzg4844.BlobToCommitment(&blob)
			if err != nil {
				return nil, err
			}
			if got != commitment {
				gotStr, _ := got.MarshalText()
				wantStr, _ := commitment.MarshalText()
				return nil, fmt.Errorf("commitment mismatch: got %v, want %v", string(gotStr), string(wantStr))
			}
		case ProofOfEquivalence:
			panic("unsupported blob proof type")
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
				BaseFeeConfig:    g.Taiko.BlockProposed.BaseFeeConfig(),
			},
		}
	case PacayaHardFork:
		panic("unsupported hardfork")
	default:
		return nil, fmt.Errorf("unsupported hardfork: %v", g.Taiko.BlockProposed.HardFork())
	}

	verifierAddress, err := g.ChainSpec.getForkVerifierAddress(g.Taiko.BlockProposed.BlockNumber(), proofType)
	if err != nil {
		return nil, err
	}

	pi := &publicInputs{
		transition: &ontake.TaikoDataTransition{
			ParentHash: g.ParentHeader.Hash(),
			BlockHash:  g.Block.Hash(),
			StateRoot:  g.Block.Root(),
			Graffiti:   g.Taiko.ProverData.Graffiti,
		},
		block_metadata: metadata,
		verifier:       verifierAddress,
		prover:         g.Taiko.ProverData.Prover,
		sgxInstance:    common.Address{},
		chainID:        g.ChainSpec.ChainId,
	}

	if g.ChainSpec.IsTaiko && !reflect.DeepEqual(pi.block_metadata, g.Taiko.BlockProposed) {
		return nil, fmt.Errorf("block hash mismatch, expected: %+v, got: %+v", g.Taiko.BlockProposed, pi.block_metadata)
	}
	return pi, nil
}

func (p *publicInputs) hash() (common.Address, error) {
	b, err := publicInputsType.Pack("VERIFY_PROOF", p.chainID, p.verifier, p.transition, p.sgxInstance, p.block_metadata.Hash())
	if err != nil {
		return common.Address{}, err
	}
	return common.Address(keccak(b)), nil
}
