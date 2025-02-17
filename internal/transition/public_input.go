package transition

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

type PublicInput struct {
	transition     *ontake.TaikoDataTransition
	block_metadata BlockMetaDataFork
	verifier       common.Address
	prover         common.Address
	sgxInstance    common.Address
	chainID        uint64
}

func (p *PublicInput) Hash() (common.Address, error) {
	b, err := publicInputsType.Pack("VERIFY_PROOF", p.chainID, p.verifier, p.transition, p.sgxInstance, p.block_metadata.Hash())
	if err != nil {
		return common.Address{}, err
	}
	return common.Address(keccak(b)), nil
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

func NewPublicInput(driver Driver, proofType ProofType) (*PublicInput, error) {
	verifierAddress, err := driver.GetForkVerifierAddress(proofType)
	if err != nil {
		return nil, err
	}

	metaData, err := driver.BlockMetaDataFork(proofType)
	if err != nil {
		return nil, err
	}

	pi := &PublicInput{
		transition:     driver.Transition(),
		block_metadata: metaData,
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
