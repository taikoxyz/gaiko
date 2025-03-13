package witness

import (
	"iter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type Pair struct {
	Input *GuestInput
	Txs   types.Transactions
}

// WitnessInput is an interface for witnesses.
type WitnessInput interface {
	// GuestInputs returns a sequence of pairs of GuestInput and Transactions.
	GuestInputs() iter.Seq[*Pair]
	// BlockProposedFork returns the block proposed data.
	BlockProposedFork() BlockProposedFork
	// BlockMetadataFork returns the block metadata.
	BlockMetadataFork() (BlockMetadataFork, error)
	// Verify verifies the witness.
	Verify(proofType ProofType) error
	// Transition returns the transition data.
	Transition() any
	// ForkVerifierAddress returns the verifier address.
	ForkVerifierAddress(proofType ProofType) (common.Address, error)
	// Prover returns the prover address.
	Prover() common.Address
	// ChainID returns the chain ID.
	ChainID() uint64
	// IsTaiko returns true if the driver is for Taiko.
	IsTaiko() bool
	// ChainConfig returns the chain config.
	ChainConfig() (*params.ChainConfig, error)
}
