package witness

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/pkg/keccak"
)

// addressToB256 converts an Address to B256 format (32 bytes with leading zeros)
func addressToB256(addr common.Address) common.Hash {
	var b [32]byte
	copy(b[12:], addr.Bytes())
	return common.BytesToHash(b[:])
}

// hashTwoValues computes keccak256(abi.encode(v0, v1))
func hashTwoValues(v0, v1 common.Hash) common.Hash {
	data := make([]byte, 64)
	copy(data[0:32], v0.Bytes())
	copy(data[32:64], v1.Bytes())
	return keccak.Keccak(data)
}

// hashThreeValues computes keccak256(abi.encode(v0, v1, v2))
func hashThreeValues(v0, v1, v2 common.Hash) common.Hash {
	data := make([]byte, 96)
	copy(data[0:32], v0.Bytes())
	copy(data[32:64], v1.Bytes())
	copy(data[64:96], v2.Bytes())
	return keccak.Keccak(data)
}

// hashFiveValues computes keccak256(abi.encode(v0, v1, v2, v3, v4))
func hashFiveValues(v0, v1, v2, v3, v4 common.Hash) common.Hash {
	data := make([]byte, 160)
	copy(data[0:32], v0.Bytes())
	copy(data[32:64], v1.Bytes())
	copy(data[64:96], v2.Bytes())
	copy(data[96:128], v3.Bytes())
	copy(data[128:160], v4.Bytes())
	return keccak.Keccak(data)
}

// uint64ToB256 converts a uint64 to B256 format
func uint64ToB256(v uint64) common.Hash {
	var b [32]byte
	binary.BigEndian.PutUint64(b[24:32], v)
	return common.BytesToHash(b[:])
}

// hashCheckpoint computes the hash of a checkpoint
// Equivalent to: keccak256(abi.encode(blockNumber, blockHash, stateRoot))
func hashCheckpoint(checkpoint *ShastaCheckpoint) common.Hash {
	return hashThreeValues(
		uint64ToB256(checkpoint.BlockNumber),
		checkpoint.BlockHash,
		checkpoint.StateRoot,
	)
}

// hashTransitionWithMetadata computes the hash of a transition with its metadata
// Equivalent to Raiko's hash_transition_with_metadata function
func hashTransitionWithMetadata(transition *ShastaTransition, metadata *ShastaTransitionMetadata) common.Hash {
	designatedProverB256 := addressToB256(metadata.DesignatedProver)
	proverB256 := addressToB256(metadata.ActualProver)

	return hashFiveValues(
		transition.ProposalHash,
		transition.ParentTransitionHash,
		hashCheckpoint(&transition.Checkpoint),
		designatedProverB256,
		proverB256,
	)
}

// HashShastaAggregation computes the public input hash for a Shasta aggregation
func HashShastaAggregation(transitionHashes []common.Hash, chainID uint64, verifierAddress common.Address, sgxInstance common.Address) common.Hash {
	return hashPublicInput(
		hashTransitionsHashArray(transitionHashes),
		chainID,
		verifierAddress,
		sgxInstance,
	)
}

// hashTransitionsHashArray computes the hash of an array of transition hashes
// Equivalent to Raiko's hash_transitions_hash_array_with_metadata function
func hashTransitionsHashArray(transitionHashes []common.Hash) common.Hash {
	if len(transitionHashes) == 0 {
		// Empty hash
		return keccak.Keccak(nil)
	}

	// For 1 transition
	if len(transitionHashes) == 1 {
		return hashTwoValues(
			uint64ToB256(uint64(len(transitionHashes))),
			transitionHashes[0],
		)
	}

	// For 2 transitions
	if len(transitionHashes) == 2 {
		return hashThreeValues(
			uint64ToB256(uint64(len(transitionHashes))),
			transitionHashes[0],
			transitionHashes[1],
		)
	}

	// For larger arrays, encode as: length || hash1 || hash2 || ...
	bufferSize := 32 + (len(transitionHashes) * 32)
	buffer := make([]byte, bufferSize)

	// Write array length
	lengthBytes := new(big.Int).SetUint64(uint64(len(transitionHashes))).FillBytes(make([]byte, 32))
	copy(buffer[0:32], lengthBytes)

	// Write each transition hash
	for i, hash := range transitionHashes {
		copy(buffer[32+(i*32):32+((i+1)*32)], hash.Bytes())
	}

	return keccak.Keccak(buffer)
}

var verifyProofB256 = common.HexToHash("0x" + "5645524946595f50524f4f46000000000000000000000000000000000000000000") // "VERIFY_PROOF" in hex

// hashFourValues computes keccak256(abi.encode(v0, v1, v2, v3))
func hashFourValues(v0, v1, v2, v3 common.Hash) common.Hash {
	data := make([]byte, 128)
	copy(data[0:32], v0.Bytes())
	copy(data[32:64], v1.Bytes())
	copy(data[64:96], v2.Bytes())
	copy(data[96:128], v3.Bytes())
	return keccak.Keccak(data)
}

// hashProposal computes the hash of a Shasta proposal
// Follows Raiko's implementation: pack 3 fields into one U256, then hash 4 values
func hashProposal(proposal *ShastaProposal) common.Hash {
	// Pack id, timestamp, and endOfSubmissionWindowTimestamp into one U256
	// packed = (id << 208) | (timestamp << 160) | (endOfSubmissionWindowTimestamp << 112)
	packed := new(big.Int)

	// id << 208
	idShifted := new(big.Int).Lsh(new(big.Int).SetUint64(proposal.ID), 208)
	packed.Or(packed, idShifted)

	// timestamp << 160
	timestampShifted := new(big.Int).Lsh(new(big.Int).SetUint64(proposal.Timestamp), 160)
	packed.Or(packed, timestampShifted)

	// endOfSubmissionWindowTimestamp << 112
	endTimeShifted := new(big.Int).Lsh(new(big.Int).SetUint64(proposal.EndOfSubmissionWindowTimestamp), 112)
	packed.Or(packed, endTimeShifted)

	// Convert to B256
	packedBytes := make([]byte, 32)
	packed.FillBytes(packedBytes)

	// Hash 4 values: packed, proposer, coreStateHash, derivationHash
	return hashFourValues(
		common.BytesToHash(packedBytes),
		addressToB256(proposal.Proposer),
		proposal.CoreStateHash,
		proposal.DerivationHash,
	)
}

// hashPublicInput computes the final public input hash for Shasta
// Equivalent to Raiko's hash_public_input function
func hashPublicInput(
	aggregatedProvingHash common.Hash,
	chainID uint64,
	verifierAddress common.Address,
	sgxInstance common.Address,
) common.Hash {
	return hashFiveValues(
		verifyProofB256,
		uint64ToB256(chainID),
		addressToB256(verifierAddress),
		aggregatedProvingHash,
		addressToB256(sgxInstance),
	)
}
