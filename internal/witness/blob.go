package witness

import (
	"errors"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

const (
	proofSize      = 48
	commitmentSize = 48
)

// verifyBlob verifies the integrity of a blob based on the provided proof type.
// It supports different types of blob proofs, such as KZG commitments and proofs of equivalence.
//
// Parameters:
// - blobProofType: The type of proof used to verify the blob.
// - blob: The blob to be verified.
// - commitment: The expected commitment for the blob.
// - proof: The proof used for verification (optional, depending on the proof type).
//
// Returns:
// - An error if the verification fails or if the proof type is unsupported, otherwise nil.
func verifyBlob(
	blobProofType BlobProofType,
	blob *eth.Blob,
	commitment kzg4844.Commitment,
	proof *kzg4844.Proof,
) error {
	switch blobProofType {
	case KzgVersionedHash:
		got, err := blob.ComputeKZGCommitment()
		if err != nil {
			return err
		}
		if got != commitment {
			gotStr, _ := got.MarshalText()
			wantStr, _ := commitment.MarshalText()
			return fmt.Errorf(
				"commitment mismatch: got %s, want %s",
				string(gotStr),
				string(wantStr),
			)
		}
	case ProofOfEquivalence:
		if proof == nil {
			return errors.New("missing proof")
		}
		return eth.VerifyBlobProof(blob, commitment, *proof)
	default:
		return fmt.Errorf("unsupported blob proof type: %s", blobProofType)
	}
	return nil
}

func getBlobProofType(proofType ProofType, blobProofTypeHint BlobProofType) BlobProofType {
	switch proofType {
	case NativeProofType:
		return blobProofTypeHint
	case SGXProofType, SGXGethProofType:
		return KzgVersionedHash
	case Sp1ProofType, Risc0ProofType:
		//TODO: Implement support for zk proofs
		return ProofOfEquivalence
	default:
		panic("unreachable")
	}
}
