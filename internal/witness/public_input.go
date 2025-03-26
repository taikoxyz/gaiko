package witness

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

type PublicInput struct {
	transition     any
	block_metadata BlockMetadataFork
	verifier       common.Address
	prover         common.Address
	sgxInstance    common.Address
	chainID        uint64
}

func (p *PublicInput) Hash() (common.Hash, error) {
	var (
		data []byte
		err  error
	)
	metaHash := p.block_metadata.Hash()
	switch transition := p.transition.(type) {
	case *ontake.TaikoDataTransition:
		data, err = publicInputsV1Type.Pack(
			"VERIFY_PROOF",
			p.chainID,
			p.verifier,
			transition,
			p.sgxInstance,
			p.prover,
			metaHash,
		)
	case *pacaya.ITaikoInboxTransition:
		data, err = publicInputsV2Type.Pack(
			"VERIFY_PROOF",
			p.chainID,
			p.verifier,
			transition,
			p.sgxInstance,
			metaHash,
		)
	default:
		return common.Hash{}, fmt.Errorf("unsupported transition type: %T", transition)
	}

	if err != nil {
		return common.Hash{}, err
	}
	return keccak.Keccak(data), nil
}

func NewPublicInput(
	input WitnessInput,
	proofType ProofType,
	sgxInstance common.Address,
) (*PublicInput, error) {
	verifier := input.ForkVerifierAddress(proofType)
	if err := input.Verify(proofType); err != nil {
		return nil, err
	}

	meta, err := input.BlockMetadataFork()
	if err != nil {
		return nil, err
	}

	pi := &PublicInput{
		transition:     input.Transition(),
		block_metadata: meta,
		verifier:       verifier,
		prover:         input.Prover(),
		sgxInstance:    common.Address{},
		chainID:        input.ChainID(),
	}

	if input.IsTaiko() && input.BlockProposedFork().BlockMetadataFork() != nil {
		got, err := meta.ABIEncode()
		if err != nil {
			return nil, err
		}
		want, err := input.BlockProposedFork().BlockMetadataFork().ABIEncode()
		if err != nil {
			return nil, err
		}
		if !slices.Equal(got, want) {
			return nil, fmt.Errorf("block hash mismatch, expected: %#x, got: %#x", want, got)
		}
	}
	return pi, nil
}
