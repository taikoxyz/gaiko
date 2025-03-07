package witness

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/internal/keccak"
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
	switch trans := p.transition.(type) {
	case *ontake.TaikoDataTransition:
		data, err = publicInputsType.Pack(
			"VERIFY_PROOF",
			p.chainID,
			p.verifier,
			trans,
			p.sgxInstance,
			p.block_metadata.Hash(),
		)
	case *pacaya.ITaikoInboxTransition:
		data, err = publicInputsType.Pack(
			"VERIFY_PROOF",
			p.chainID,
			p.verifier,
			trans,
			p.sgxInstance,
			p.block_metadata.Hash(),
		)
	}

	if err != nil {
		return common.Hash{}, err
	}
	return keccak.Keccak(data), nil
}

func NewPublicInput(
	driver GuestDriver,
	proofType ProofType,
	sgxInstance common.Address,
) (*PublicInput, error) {
	verifierAddress, err := driver.ForkVerifierAddress(proofType)
	if err != nil {
		return nil, err
	}

	meta, err := driver.BlockMetadataFork(proofType)
	if err != nil {
		return nil, err
	}

	pi := &PublicInput{
		transition:     driver.Transition(),
		block_metadata: meta,
		verifier:       verifierAddress,
		prover:         driver.Prover(),
		sgxInstance:    common.Address{},
		chainID:        driver.ChainID(),
	}

	if driver.IsTaiko() {
		got, _ := pi.block_metadata.ABIEncode()
		want, _ := driver.BlockProposedFork().ABIEncode()
		if !slices.Equal(got, want) {
			return nil, fmt.Errorf("block hash mismatch, expected: %#x, got: %#x", want, got)
		}
	}
	return pi, nil
}
