package transition

import (
	"iter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
)

type Pair struct {
	*GuestInput
	types.Transactions
}

type GuestDriver interface {
	GuestInputs() iter.Seq[Pair]
	BlockProposedFork() BlockProposedFork
	BlockMetaDataFork(proofType ProofType) (BlockMetaDataFork, error)
	Transition() *ontake.TaikoDataTransition
	ForkVerifierAddress(proofType ProofType) (common.Address, error)
	Prover() common.Address
	ChainID() uint64
	IsTaiko() bool
	ChainConfig() (*params.ChainConfig, error)
}
