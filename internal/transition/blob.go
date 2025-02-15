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

func (g *BatchGuestInput) verifyBatchModeBlobUsage(proofType ProofType) error {
	blobProofType := getBlobProofType(proofType, g.Taiko.BlobProofType)
	for i := 0; i < len(g.Taiko.TxDataFromBlob); i++ {
		blob := g.Taiko.TxDataFromBlob[i]
		commitment := (*g.Taiko.BlobCommitments)[i]
		proof := (*g.Taiko.BlobProofs)[i]
		if err := verifyBlob(blobProofType, blob, commitment, (*kzg4844.Proof)(&proof)); err != nil {
			return err
		}
	}
	return nil
}

func verifyBlob(blobProofType BlobProofType, blob eth.Blob, commitment kzg4844.Commitment, proof *kzg4844.Proof) error {
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
		return eth.VerifyBlobProof(&blob, commitment, *proof)
	default:
		return fmt.Errorf("unsupported blob proof type: %v", blobProofType)
	}
	return nil
}
