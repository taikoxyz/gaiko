package transition

import (
	"iter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

type Pair struct {
	*GuestInput
	types.Transactions
}

type Driver interface {
	GuestInputs() iter.Seq[Pair]
	BlockProposedFork() BlockProposedFork
	BlockMetaDataFork(proofType ProofType) (BlockMetaDataFork, error)
	Transition() *ontake.TaikoDataTransition
	GetForkVerifierAddress(proofType ProofType) (common.Address, error)
	Prover() common.Address
	chainID() uint64
	IsTaiko() bool
}
