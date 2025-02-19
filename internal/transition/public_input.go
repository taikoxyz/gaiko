package transition

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/internal"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

type PublicInput struct {
	transition     *ontake.TaikoDataTransition
	block_metadata BlockMetadataFork
	verifier       common.Address
	prover         common.Address
	sgxInstance    common.Address
	chainID        uint64
}

func (p *PublicInput) Hash() (common.Hash, error) {
	b, err := publicInputsType.Pack("VERIFY_PROOF", p.chainID, p.verifier, p.transition, p.sgxInstance, p.block_metadata.Hash())
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(internal.Keccak(b)), nil
}

func NewPublicInput(driver GuestDriver, proofType ProofType, sgxInstance common.Address) (*PublicInput, error) {
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
		got, _ := pi.block_metadata.Encode()
		want, _ := driver.BlockProposedFork().Encode()
		if !slices.Equal(got, want) {
			return nil, fmt.Errorf("block hash mismatch, expected: %+v, got: %+v", want, got)
		}
	}
	return pi, nil
}
