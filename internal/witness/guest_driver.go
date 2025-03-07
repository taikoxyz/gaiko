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

// GuestDriver is an interface for guest inputs.
type GuestDriver interface {
	// GuestInputs returns a sequence of pairs of GuestInput and Transactions.
	GuestInputs() iter.Seq[*Pair]
	// BlockProposedFork returns the block proposed data.
	BlockProposedFork() BlockProposedFork
	// BlockMetadataFork returns the block metadata.
	BlockMetadataFork(proofType ProofType) (BlockMetadataFork, error)
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
