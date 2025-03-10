package witness

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/pkg/keccak"
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
	data, err := publicInputsType.Pack(
		"VERIFY_PROOF",
		p.chainID,
		p.verifier,
		p.transition,
		p.sgxInstance,
		p.block_metadata.Hash(),
	)
	if err != nil {
		return common.Hash{}, err
	}
	return keccak.Keccak(data), nil
}

func NewPublicInput(
	wit Witness,
	proofType ProofType,
	sgxInstance common.Address,
) (*PublicInput, error) {
	verifierAddress, err := wit.ForkVerifierAddress(proofType)
	if err != nil {
		return nil, err
	}

	if err := wit.Verify(proofType); err != nil {
		return nil, err
	}

	meta, err := wit.BlockMetadataFork()
	if err != nil {
		return nil, err
	}

	pi := &PublicInput{
		transition:     wit.Transition(),
		block_metadata: meta,
		verifier:       verifierAddress,
		prover:         wit.Prover(),
		sgxInstance:    common.Address{},
		chainID:        wit.ChainID(),
	}

	if wit.IsTaiko() && wit.BlockProposedFork().BlockMetadataFork() != nil {
		got, _ := pi.block_metadata.ABIEncode()
		want, _ := wit.BlockProposedFork().BlockMetadataFork().ABIEncode()
		if !slices.Equal(got, want) {
			return nil, fmt.Errorf("block hash mismatch, expected: %#x, got: %#x", want, got)
		}
	}
	return pi, nil
}
