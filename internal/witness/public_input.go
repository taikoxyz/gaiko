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
	transition    any
	blockMetadata BlockMetadata
	blockProposed BlockProposed
	verifier      common.Address
	prover        common.Address
	sgxInstance   common.Address
	chainID       uint64
}

func (p *PublicInput) Hash() (common.Hash, error) {
	// Shasta uses a different hash calculation
	if p.blockProposed.IsShasta() {
		transitionHash, ok := p.transition.(common.Hash)
		if !ok {
			return common.Hash{}, fmt.Errorf("shasta transition must be []common.Hash, got %T", p.transition)
		}
		return transitionHash, nil
	}

	var (
		data []byte
		err  error
	)
	metaHash := p.blockMetadata.Hash()
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
	input GuestInput,
	proofType ProofType,
	sgxType string,
	sgxInstance common.Address,
) (*PublicInput, error) {
	verifier := input.ForkVerifierAddress(proofType)
	// ignore verify in debug/test mode, the specs changed, only check the block hash
	if sgxType != "debug" {
		if err := input.Verify(proofType); err != nil {
			return nil, err
		}
	}

	meta, err := input.BlockMetadata()
	if err != nil {
		return nil, err
	}

	// Check if this is Shasta

	pi := &PublicInput{
		transition:    input.Transition(),
		blockMetadata: meta,
		blockProposed: input.BlockProposed(),
		verifier:      verifier,
		prover:        input.Prover(),
		sgxInstance:   sgxInstance,
		chainID:       input.ChainID(),
	}

	if input.IsTaiko() && input.BlockProposed().BlockMetadata() != nil {
		got, err := meta.ABIEncode()
		if err != nil {
			return nil, err
		}
		want, err := input.BlockProposed().BlockMetadata().ABIEncode()
		if err != nil {
			return nil, err
		}
		if !slices.Equal(got, want) {
			return nil, fmt.Errorf("metadata hash mismatch, expected: %#x, got: %#x", want, got)
		}
	}
	return pi, nil
}
