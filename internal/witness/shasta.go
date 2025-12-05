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

// ShastaProposalCheckpoint represents a checkpoint in the Shasta proof proposal.
type ShastaProposalCheckpoint struct {
	BlockNumber uint64      `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	StateRoot   common.Hash `json:"stateRoot"`
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
