package transition

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

const (
	proofSize      = 48
	commitmentSize = 48
)

func verifyBlob(blobProofType BlobProofType, blob *eth.Blob, commitment kzg4844.Commitment, proof *kzg4844.Proof) error {
	switch blobProofType {
	case KzgVersionedHash:
		got, err := blob.ComputeKZGCommitment()
		if err != nil {
			return err
		}
		if got != commitment {
			gotStr, _ := got.MarshalText()
			wantStr, _ := commitment.MarshalText()
			return fmt.Errorf("commitment mismatch: got %v, want %v", string(gotStr), string(wantStr))
		}
	case ProofOfEquivalence:
		if proof == nil {
			return fmt.Errorf("missing proof")
		}
		return eth.VerifyBlobProof(blob, commitment, *proof)
	default:
		return fmt.Errorf("unsupported blob proof type: %v", blobProofType)
	}
	return nil
}

func getBlobProofType(proofType ProofType, blobProofTypeHint BlobProofType) BlobProofType {
	switch proofType {
	case NativeProofType:
		return blobProofTypeHint
	case SgxProofType, GaikoSgxProofType:
		return KzgVersionedHash
	case Sp1ProofType, Risc0ProofType:
		return ProofOfEquivalence
	default:
		panic("unreachable")
	}
}
