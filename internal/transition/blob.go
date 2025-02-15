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
		_blob := g.Taiko.TxDataFromBlob[i]
		_commitment := (*g.Taiko.BlobCommitments)[i]
		_proof := (*g.Taiko.BlobProofs)[i]
		if err := verifyBlob(blobProofType, _blob, _commitment, &_proof); err != nil {
			return err
		}
	}
	return nil
}

func verifyBlob(
	blobProofType BlobProofType,
	_blob [eth.BlobSize]byte,
	_commitment [commitmentSize]byte,
	_proof *[proofSize]byte) error {
	commitment := kzg4844.Commitment(_commitment)
	blob := eth.Blob(_blob)
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
		if _proof == nil {
			return fmt.Errorf("missing proof")
		}
		proof := kzg4844.Proof(*_proof)
		return eth.VerifyBlobProof(&blob, commitment, proof)
	default:
		return fmt.Errorf("unsupported blob proof type: %v", blobProofType)
	}
	return nil
}
