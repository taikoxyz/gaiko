package witness

import (
	"iter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type Pair struct {
	Input *SingleGuestInput
	Txs   types.Transactions
}

type ID struct {
	BatchID uint64
	BlockID uint64
}

// GuestInput is an interface for witnesses.
type GuestInput interface {
	// GuestInputs returns a sequence of pairs of GuestInput and Transactions.
	GuestInputs() iter.Seq[*Pair]
	// BlockProposed returns the block proposed data.
	BlockProposed() BlockProposed
	// BlockMetadata returns the block metadata.
	BlockMetadata() (BlockMetadata, error)
	// Verify verifies the witness.
	Verify(proofType ProofType) error
	// Transition returns the transition data.
	Transition() any
	// ForkVerifierAddress returns the verifier address.
	ForkVerifierAddress(proofType ProofType) common.Address
	// Prover returns the prover address.
	Prover() common.Address
	// ChainID returns the chain ID.
	ChainID() uint64
	// Block ID or Batch ID
	ID() ID
	// IsTaiko returns true if the driver is for Taiko.
	IsTaiko() bool
	// ChainConfig returns the chain config.
	ChainConfig() (*params.ChainConfig, error)
}
