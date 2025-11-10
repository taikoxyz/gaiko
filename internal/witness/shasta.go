package witness

import "github.com/ethereum/go-ethereum/common"

type ShastaBlobSlice struct {
	BlobHashes []common.Hash
	Offset     uint32
	Timestamp  uint64
}

type ShastaDerivationSource struct {
	IsForcedInclusion bool
	BlobSlice         ShastaBlobSlice
}

type ShastaDerivation struct {
	OriginBlockNumber  uint64
	OriginBlockHash    common.Hash
	BasefeeSharingPctg uint8
	Sources            []ShastaDerivationSource
}

type ShastaProposal struct {
	ID                             uint64
	Timestamp                      uint64
	EndOfSubmissionWindowTimestamp uint64
	Proposer                       common.Address
	CoreStateHash                  common.Hash
	DerivationHash                 common.Hash
}

type ShastaCoreState struct {
	NextProposalID              uint64
	LastProposalBlockID         uint64
	LastFinalizedProposalID     uint64
	LastCheckpointTimestamp     uint64
	LastFinalizedTransitionHash common.Hash
	BondInstructionsHash        common.Hash
}

type ShastaEventData struct {
	Proposal   ShastaProposal
	Derivation ShastaDerivation
	CoreState  ShastaCoreState
}

// ShastaProofProposal represents a Shasta proposal for proving (different from ShastaProposal event).
type ShastaProofProposal struct {
	ProposalID             uint64                   `json:"proposal_id"`
	DesignatedProver       common.Address           `json:"designated_prover"`
	ParentTransitionHash   common.Hash              `json:"parent_transition_hash"`
	Checkpoint             ShastaProposalCheckpoint `json:"checkpoint"`
	L1InclusionBlockNumber uint64                   `json:"l1_inclusion_block_number"`
	L2BlockNumbers         []uint64                 `json:"l2_block_numbers"`
}

// ShastaProposalCheckpoint represents a checkpoint in the Shasta proof proposal.
type ShastaProposalCheckpoint struct {
	BlockNumber uint64      `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	StateRoot   common.Hash `json:"stateRoot"`
}

// ShastaAggregationGuestInput represents the input for Shasta aggregation proofs.
type ShastaAggregationGuestInput struct {
	Proposals []*ShastaProofProposal `json:"proposals"`
}

// ShastaCheckpoint represents a checkpoint in a Shasta transition
type ShastaCheckpoint struct {
	BlockNumber uint64
	BlockHash   common.Hash
	StateRoot   common.Hash
}

// ShastaTransition represents a state transition in Shasta
type ShastaTransition struct {
	ProposalHash         common.Hash
	ParentTransitionHash common.Hash
	Checkpoint           ShastaCheckpoint
}

// ShastaTransitionMetadata contains metadata for a Shasta transition
type ShastaTransitionMetadata struct {
	DesignatedProver common.Address
	ActualProver     common.Address
}
