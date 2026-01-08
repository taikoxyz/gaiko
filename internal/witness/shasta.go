package witness

import "github.com/ethereum/go-ethereum/common"

type ShastaBlobSlice struct {
	BlobHashes []common.Hash `json:"blobHashes"`
	Offset     uint32        `json:"offset"`
	Timestamp  uint64        `json:"timestamp"`
}

type ShastaDerivationSource struct {
	IsForcedInclusion bool            `json:"isForcedInclusion"`
	BlobSlice         ShastaBlobSlice `json:"blobSlice"`
}

type ShastaProposal struct {
	ID                             uint64                   `json:"id"`
	Timestamp                      uint64                   `json:"timestamp"`
	EndOfSubmissionWindowTimestamp uint64                   `json:"endOfSubmissionWindowTimestamp"`
	Proposer                       common.Address           `json:"proposer"`
	ParentProposalHash             common.Hash              `json:"parentProposalHash"`
	OriginBlockNumber              uint64                   `json:"originBlockNumber"`
	OriginBlockHash                common.Hash              `json:"originBlockHash"`
	BasefeeSharingPctg             uint8                    `json:"basefeeSharingPctg"`
	Sources                        []ShastaDerivationSource `json:"sources"`
}

type ShastaEventData struct {
	Proposal ShastaProposal `json:"proposal"`
}

// ShastaProposalCheckpoint represents a checkpoint in the Shasta proof proposal.
type ShastaProposalCheckpoint struct {
	BlockNumber uint64      `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	StateRoot   common.Hash `json:"stateRoot"`
}

// ShastaCheckpoint represents a checkpoint in a Shasta transition
type ShastaCheckpoint struct {
	BlockNumber uint64      `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	StateRoot   common.Hash `json:"stateRoot"`
}

// ShastaTransitionInput represents the fields signed by sub-proofs in Shasta.
type ShastaTransitionInput struct {
	Proposer  common.Address `json:"proposer"`
	Timestamp uint64         `json:"timestamp"`
}

// TransitionInputData is the full Shasta transition input used for hashing.
type TransitionInputData struct {
	ProposalID         uint64                `json:"proposal_id"`
	ProposalHash       common.Hash           `json:"proposal_hash"`
	ParentProposalHash common.Hash           `json:"parent_proposal_hash"`
	ParentBlockHash    common.Hash           `json:"parent_block_hash"`
	ActualProver       common.Address        `json:"actual_prover"`
	Transition         ShastaTransitionInput `json:"transition"`
	Checkpoint         ShastaCheckpoint      `json:"checkpoint"`
}

// ProofCarryData is attached to Shasta proofs for aggregation.
type ProofCarryData struct {
	ChainID         uint64              `json:"chain_id"`
	Verifier        common.Address      `json:"verifier"`
	TransitionInput TransitionInputData `json:"transition_input"`
}

// ShastaTransition represents a proposal transition in commitment hashing.
type ShastaTransition struct {
	Proposer  common.Address `json:"proposer"`
	Timestamp uint64         `json:"timestamp"`
	BlockHash common.Hash    `json:"blockHash"`
}

// ShastaCommitment is the commitment data hashed for aggregation.
type ShastaCommitment struct {
	FirstProposalID              uint64             `json:"firstProposalId"`
	FirstProposalParentBlockHash common.Hash        `json:"firstProposalParentBlockHash"`
	LastProposalHash             common.Hash        `json:"lastProposalHash"`
	ActualProver                 common.Address     `json:"actualProver"`
	EndBlockNumber               uint64             `json:"endBlockNumber"`
	EndStateRoot                 common.Hash        `json:"endStateRoot"`
	Transitions                  []ShastaTransition `json:"transitions"`
}
