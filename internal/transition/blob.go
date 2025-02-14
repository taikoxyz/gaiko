package transition

import (
	"crypto/sha256"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

const (
	blobSize       = 131072
	proofSize      = 48
	commitmentSize = 48
)

func (g *BatchGuestInput) verifyBatchModeBlobUsage(proofType ProofType) error {
	blobProofType := getBlobProofType(proofType, g.Taiko.BlobProofType)
	for i := 0; i < len(g.Taiko.TxDataFromBlob); i++ {
		blobData := g.Taiko.TxDataFromBlob[i]
		_commitment := (*g.Taiko.BlobCommitments)[i]
		_proof := (*g.Taiko.BlobProofs)[i]
		commitment := kzg4844.Commitment(_commitment)
		versionedHash := common.Hash(kzg4844.CalcBlobHashV1(sha256.New(), &commitment))
		if err := verifyBlob(blobProofType, blobData, versionedHash, _commitment, &_proof); err != nil {
			return err
		}
	}
	return nil
}

func verifyBlob(
	blobProofType BlobProofType,
	_blob [131072]byte,
	versionedHash common.Hash,
	_commitment [48]byte,
	_proof *[48]byte) error {
	commitment := kzg4844.Commitment(_commitment)
	blob := kzg4844.Blob(_blob)
	switch blobProofType {
	case KzgVersionedHash:
		got, err := kzg4844.BlobToCommitment(&blob)
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
		return kzg4844.VerifyBlobProof(&blob, commitment, proof)
	default:
		return fmt.Errorf("unsupported blob proof type: %v", blobProofType)
	}
	return nil
}
