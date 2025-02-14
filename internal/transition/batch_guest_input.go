package transition

import "github.com/ethereum/go-ethereum/core/types"

type BatchGuestInput struct {
	inputs []GuestInput
}

type TaikoGuestBatchInput struct {
	BatchId            uint64
	L1Header           types.Header
	BatchProposed      BlockProposedFork
	ChainSpec          ChainSpec
	ProverData         TaikoProverData
	TxDataFromCalldata []byte
	TxDataFromBlob     [][]byte
	BlobCommitments    *[][]byte
	BlobProofs         *[][]byte
	BlobProofType      BlobProofType
}
