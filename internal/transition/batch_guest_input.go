package transition

import "github.com/ethereum/go-ethereum/core/types"

type BatchGuestInput struct {
	Inputs []GuestInput
	Taiko  TaikoGuestBatchInput
}

type TaikoGuestBatchInput struct {
	BatchId            uint64
	L1Header           types.Header
	BatchProposed      BlockProposedFork
	ChainSpec          ChainSpec
	ProverData         TaikoProverData
	TxDataFromCalldata []byte
	TxDataFromBlob     [][blobSize]byte
	BlobCommitments    *[][commitmentSize]byte
	BlobProofs         *[][proofSize]byte
	BlobProofType      BlobProofType
}

func (g *BatchGuestInput) publicInputs(proofType ProofType) (*publicInput, error) {
	panic("unimplemented")
}
