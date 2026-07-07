package witness

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/pkg/keccak"
)

var (
	shastaBlobSliceComponents = []abi.ArgumentMarshaling{
		{Name: "blobHashes", Type: "bytes32[]"},
		{Name: "offset", Type: "uint24"},
		{Name: "timestamp", Type: "uint48"},
	}
	shastaDerivationSourceComponents = []abi.ArgumentMarshaling{
		{Name: "isForcedInclusion", Type: "bool"},
		{Name: "blobSlice", Type: "tuple", Components: shastaBlobSliceComponents},
	}
	shastaProposalComponents = []abi.ArgumentMarshaling{
		{Name: "id", Type: "uint48"},
		{Name: "timestamp", Type: "uint48"},
		{Name: "endOfSubmissionWindowTimestamp", Type: "uint48"},
		{Name: "proposer", Type: "address"},
		{Name: "parentProposalHash", Type: "bytes32"},
		{Name: "originBlockNumber", Type: "uint48"},
		{Name: "originBlockHash", Type: "bytes32"},
		{Name: "basefeeSharingPctg", Type: "uint8"},
		{Name: "sources", Type: "tuple[]", Components: shastaDerivationSourceComponents},
	}
	shastaProposalType, _ = abi.NewType("tuple", "Proposal", shastaProposalComponents)
	shastaProposalArgs    = abi.Arguments{{Name: "proposal", Type: shastaProposalType}}
)

var verifyProofB256 = common.HexToHash("5645524946595f50524f4f460000000000000000000000000000000000000000") // "VERIFY_PROOF" in hex

// GetVerifyProofB256 returns the VERIFY_PROOF constant for debugging.
func GetVerifyProofB256() common.Hash {
	return verifyProofB256
}

// addressToB256 converts an Address to B256 format (32 bytes with leading zeros).
func addressToB256(addr common.Address) common.Hash {
	var b [32]byte
	copy(b[12:], addr.Bytes())
	return common.BytesToHash(b[:])
}

// AddressToB256 is exported for debugging.
func AddressToB256(addr common.Address) common.Hash {
	return addressToB256(addr)
}

// uint64ToB256 converts a uint64 to B256 format.
func uint64ToB256(v uint64) common.Hash {
	var b [32]byte
	binary.BigEndian.PutUint64(b[24:32], v)
	return common.BytesToHash(b[:])
}

// Uint64ToB256 is exported for debugging.
func Uint64ToB256(v uint64) common.Hash {
	return uint64ToB256(v)
}

// uint48ToB256 converts a uint48 (stored in uint64) to B256 format.
func uint48ToB256(v uint64) common.Hash {
	return uint64ToB256(v & 0xffffffffffff)
}

// hashValues computes keccak256(abi.encode(values...)) with values as 32-byte words.
func hashValues(values []common.Hash) common.Hash {
	if len(values) == 0 {
		return keccak.Keccak(nil)
	}
	data := make([]byte, 32*len(values))
	for i, v := range values {
		copy(data[i*32:(i+1)*32], v.Bytes())
	}
	return keccak.Keccak(data)
}

// HashValues is exported for debugging.
func HashValues(values []common.Hash) common.Hash {
	return hashValues(values)
}

// hashTwoValues computes keccak256(abi.encode(v0, v1)).
func hashTwoValues(v0, v1 common.Hash) common.Hash {
	return hashValues([]common.Hash{v0, v1})
}

// HashTwoValues is exported for testing.
func HashTwoValues(v0, v1 common.Hash) common.Hash {
	return hashTwoValues(v0, v1)
}

// hashThreeValues computes keccak256(abi.encode(v0, v1, v2)).
func hashThreeValues(v0, v1, v2 common.Hash) common.Hash {
	return hashValues([]common.Hash{v0, v1, v2})
}

// hashFourValues computes keccak256(abi.encode(v0, v1, v2, v3)).
func hashFourValues(v0, v1, v2, v3 common.Hash) common.Hash {
	return hashValues([]common.Hash{v0, v1, v2, v3})
}

// hashFiveValues computes keccak256(abi.encode(v0, v1, v2, v3, v4)).
func hashFiveValues(v0, v1, v2, v3, v4 common.Hash) common.Hash {
	return hashValues([]common.Hash{v0, v1, v2, v3, v4})
}

// hashCheckpoint computes the hash of a checkpoint.
// Equivalent to: keccak256(abi.encode(blockNumber, blockHash, stateRoot)).
func hashCheckpoint(checkpoint *ShastaCheckpoint) common.Hash {
	return hashThreeValues(
		uint64ToB256(checkpoint.BlockNumber),
		checkpoint.BlockHash,
		checkpoint.StateRoot,
	)
}

// hashShastaTransitionInput binds all continuity-critical fields for Shasta.
func hashShastaTransitionInput(transitionInput *TransitionInputData) common.Hash {
	values := make([]common.Hash, 0, 11)
	values = append(values,
		uint64ToB256(transitionInput.ProposalID),
		transitionInput.ProposalHash,
		transitionInput.ParentProposalHash,
		transitionInput.ParentBlockHash,
		addressToB256(transitionInput.ActualProver),
		addressToB256(transitionInput.Transition.Proposer),
		uint48ToB256(transitionInput.Transition.Timestamp),
		hashCheckpoint(&transitionInput.Checkpoint),
		uint48ToB256(transitionInput.Checkpoint.BlockNumber),
		transitionInput.Checkpoint.BlockHash,
		transitionInput.Checkpoint.StateRoot,
	)
	return hashValues(values)
}

// HashShastaSubproofInput computes the domain-separated public input hash for a Shasta sub-proof.
func HashShastaSubproofInput(carry *ProofCarryData) common.Hash {
	transitionHash := hashShastaTransitionInput(&carry.TransitionInput)
	return hashFourValues(
		verifyProofB256,
		uint64ToB256(carry.ChainID),
		addressToB256(carry.Verifier),
		transitionHash,
	)
}

// hashCommitment computes the hash of a Shasta commitment, matching the Solidity layout.
func hashCommitment(commitment *ShastaCommitment) common.Hash {
	transitionsLen := len(commitment.Transitions)
	buffer := make([]common.Hash, 0, 9+transitionsLen*3)

	// [0] offset to commitment (0x20)
	buffer = append(buffer, uint64ToB256(0x20))
	// [1] firstProposalId
	buffer = append(buffer, uint64ToB256(commitment.FirstProposalID))
	// [2] firstProposalParentBlockHash
	buffer = append(buffer, commitment.FirstProposalParentBlockHash)
	// [3] lastProposalHash
	buffer = append(buffer, commitment.LastProposalHash)
	// [4] actualProver
	buffer = append(buffer, addressToB256(commitment.ActualProver))
	// [5] endBlockNumber
	buffer = append(buffer, uint64ToB256(commitment.EndBlockNumber))
	// [6] endStateRoot
	buffer = append(buffer, commitment.EndStateRoot)
	// [7] offset to transitions (0xe0)
	buffer = append(buffer, uint64ToB256(0xe0))
	// [8] transitions length
	buffer = append(buffer, uint64ToB256(uint64(transitionsLen)))

	for _, transition := range commitment.Transitions {
		buffer = append(buffer,
			addressToB256(transition.Proposer),
			uint64ToB256(transition.Timestamp),
			transition.BlockHash,
		)
	}

	return hashValues(buffer)
}

// HashCommitmentDebug returns the buffer used for commitment hashing (for debugging).
func HashCommitmentDebug(commitment *ShastaCommitment) []common.Hash {
	transitionsLen := len(commitment.Transitions)
	buffer := make([]common.Hash, 0, 9+transitionsLen*3)

	buffer = append(buffer, uint64ToB256(0x20))
	buffer = append(buffer, uint64ToB256(commitment.FirstProposalID))
	buffer = append(buffer, commitment.FirstProposalParentBlockHash)
	buffer = append(buffer, commitment.LastProposalHash)
	buffer = append(buffer, addressToB256(commitment.ActualProver))
	buffer = append(buffer, uint64ToB256(commitment.EndBlockNumber))
	buffer = append(buffer, commitment.EndStateRoot)
	buffer = append(buffer, uint64ToB256(0xe0))
	buffer = append(buffer, uint64ToB256(uint64(transitionsLen)))

	for _, transition := range commitment.Transitions {
		buffer = append(buffer,
			addressToB256(transition.Proposer),
			uint64ToB256(transition.Timestamp),
			transition.BlockHash,
		)
	}

	return buffer
}

// BuildCommitmentFromProofCarryDataVec exports the commitment building for debugging.
func BuildCommitmentFromProofCarryDataVec(proofCarryDataVec []ProofCarryData) (*ShastaCommitment, bool) {
	return buildShastaCommitmentFromProofCarryDataVec(proofCarryDataVec)
}

// hashPublicInput computes keccak256(abi.encode(VERIFY_PROOF, chain_id, verifier, prove_input_hash, sgx_instance)).
func hashPublicInput(
	proveInputHash common.Hash,
	chainID uint64,
	verifierAddress common.Address,
	sgxInstance common.Address,
) common.Hash {
	return hashFiveValues(
		verifyProofB256,
		uint64ToB256(chainID),
		addressToB256(verifierAddress),
		proveInputHash,
		addressToB256(sgxInstance),
	)
}

func encodeShastaProposal(proposal *ShastaProposal) ([]byte, error) {
	if proposal == nil {
		return nil, errors.New("nil shasta proposal")
	}

	type shastaBlobSliceABI struct {
		BlobHashes []common.Hash `abi:"blobHashes"`
		Offset     *big.Int      `abi:"offset"`
		Timestamp  *big.Int      `abi:"timestamp"`
	}
	type shastaDerivationSourceABI struct {
		IsForcedInclusion bool               `abi:"isForcedInclusion"`
		BlobSlice         shastaBlobSliceABI `abi:"blobSlice"`
	}
	type shastaProposalABI struct {
		Id                             *big.Int                    `abi:"id"`
		Timestamp                      *big.Int                    `abi:"timestamp"`
		EndOfSubmissionWindowTimestamp *big.Int                    `abi:"endOfSubmissionWindowTimestamp"`
		Proposer                       common.Address              `abi:"proposer"`
		ParentProposalHash             common.Hash                 `abi:"parentProposalHash"`
		OriginBlockNumber              *big.Int                    `abi:"originBlockNumber"`
		OriginBlockHash                common.Hash                 `abi:"originBlockHash"`
		BasefeeSharingPctg             uint8                       `abi:"basefeeSharingPctg"`
		Sources                        []shastaDerivationSourceABI `abi:"sources"`
	}

	sources := make([]shastaDerivationSourceABI, 0, len(proposal.Sources))
	for _, source := range proposal.Sources {
		sources = append(sources, shastaDerivationSourceABI{
			IsForcedInclusion: source.IsForcedInclusion,
			BlobSlice: shastaBlobSliceABI{
				BlobHashes: source.BlobSlice.BlobHashes,
				Offset:     new(big.Int).SetUint64(uint64(source.BlobSlice.Offset)),
				Timestamp:  new(big.Int).SetUint64(source.BlobSlice.Timestamp),
			},
		})
	}

	encoded, err := shastaProposalArgs.Pack(shastaProposalABI{
		Id:                             new(big.Int).SetUint64(proposal.ID),
		Timestamp:                      new(big.Int).SetUint64(proposal.Timestamp),
		EndOfSubmissionWindowTimestamp: new(big.Int).SetUint64(proposal.EndOfSubmissionWindowTimestamp),
		Proposer:                       proposal.Proposer,
		ParentProposalHash:             proposal.ParentProposalHash,
		OriginBlockNumber:              new(big.Int).SetUint64(proposal.OriginBlockNumber),
		OriginBlockHash:                proposal.OriginBlockHash,
		BasefeeSharingPctg:             proposal.BasefeeSharingPctg,
		Sources:                        sources,
	})
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// hashProposal computes the hash of a Shasta proposal using ABI encoding.
func hashProposal(proposal *ShastaProposal) common.Hash {
	encoded, err := encodeShastaProposal(proposal)
	if err != nil {
		return common.Hash{}
	}
	return keccak.Keccak(encoded)
}

// ValidateShastaProofCarryDataVec ensures a proof_carry_data_vec is consistent for aggregation.
func ValidateShastaProofCarryDataVec(proofCarryDataVec []ProofCarryData) bool {
	if len(proofCarryDataVec) == 0 {
		return false
	}
	expectedProver := proofCarryDataVec[0].TransitionInput.ActualProver
	for _, item := range proofCarryDataVec {
		if item.TransitionInput.ActualProver != expectedProver {
			return false
		}
	}
	for i := 1; i < len(proofCarryDataVec); i++ {
		prev := proofCarryDataVec[i-1]
		next := proofCarryDataVec[i]
		if prev.TransitionInput.ProposalID+1 != next.TransitionInput.ProposalID {
			return false
		}
		if prev.TransitionInput.ProposalHash != next.TransitionInput.ParentProposalHash {
			return false
		}
		if prev.ChainID != next.ChainID {
			return false
		}
		if prev.Verifier != next.Verifier {
			return false
		}
		if prev.TransitionInput.Checkpoint.BlockHash != next.TransitionInput.ParentBlockHash {
			return false
		}
	}
	return true
}

func buildShastaCommitmentFromProofCarryDataVec(proofCarryDataVec []ProofCarryData) (*ShastaCommitment, bool) {
	if !ValidateShastaProofCarryDataVec(proofCarryDataVec) {
		return nil, false
	}
	last := proofCarryDataVec[len(proofCarryDataVec)-1]

	transitions := make([]ShastaTransition, 0, len(proofCarryDataVec))
	for _, item := range proofCarryDataVec {
		transitions = append(transitions, ShastaTransition{
			Proposer:  item.TransitionInput.Transition.Proposer,
			Timestamp: item.TransitionInput.Transition.Timestamp,
			BlockHash: item.TransitionInput.Checkpoint.BlockHash,
		})
	}

	return &ShastaCommitment{
		FirstProposalID:              proofCarryDataVec[0].TransitionInput.ProposalID,
		FirstProposalParentBlockHash: proofCarryDataVec[0].TransitionInput.ParentBlockHash,
		LastProposalHash:             last.TransitionInput.ProposalHash,
		ActualProver:                 proofCarryDataVec[0].TransitionInput.ActualProver,
		EndBlockNumber:               last.TransitionInput.Checkpoint.BlockNumber,
		EndStateRoot:                 last.TransitionInput.Checkpoint.StateRoot,
		Transitions:                  transitions,
	}, true
}

// ShastaPCDAggregationHash computes the aggregation hash based on proof carry data.
func ShastaPCDAggregationHash(
	proofCarryDataVec []ProofCarryData,
	sgxInstance common.Address,
) (common.Hash, error) {
	commitment, ok := buildShastaCommitmentFromProofCarryDataVec(proofCarryDataVec)
	if !ok {
		return common.Hash{}, errors.New("invalid shasta proof carry data")
	}
	first := proofCarryDataVec[0]
	commitmentHash := hashCommitment(commitment)
	return hashPublicInput(commitmentHash, first.ChainID, first.Verifier, sgxInstance), nil
}
