package witness

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	gaikoTypes "github.com/taikoxyz/gaiko/internal/types"
)

type batchGuestInputJSON struct {
	Inputs []*guestInputJSON         `json:"inputs"`
	Taiko  *taikoGuestBatchInputJSON `json:"taiko"`
}

func (g *batchGuestInputJSON) GethType() *BatchGuestInput {
	inputs := make([]*GuestInput, len(g.Inputs))
	res := &BatchGuestInput{
		Inputs: inputs,
		Taiko:  g.Taiko.GethType(),
	}
	for i, input := range g.Inputs {
		input := input.GethType()
		input.parent = res
		inputs[i] = input
	}
	return res
}

type taikoGuestBatchInputJSON struct {
	BatchID            uint64                  `json:"batch_id"`
	L1Header           *gaikoTypes.Header      `json:"l1_header"`
	BatchProposed      *blockProposedForkJSON  `json:"batch_proposed"`
	ChainSpec          *ChainSpec              `json:"chain_spec"`
	ProverData         *TaikoProverData        `json:"prover_data"`
	TxDataFromCalldata []byte                  `json:"tx_data_from_calldata"`
	TxDataFromBlob     [][eth.BlobSize]byte    `json:"tx_data_from_blob"`
	BlobCommitments    *[][commitmentSize]byte `json:"blob_commitments"`
	BlobProofs         *[][proofSize]byte      `json:"blob_proofs"`
	BlobProofType      BlobProofType           `json:"blob_proof_type"`
}

func (t *taikoGuestBatchInputJSON) GethType() *TaikoGuestBatchInput {
	return &TaikoGuestBatchInput{
		BatchID:            t.BatchID,
		L1Header:           t.L1Header.GethType(),
		BatchProposed:      t.BatchProposed.GethType(),
		ChainSpec:          t.ChainSpec,
		ProverData:         t.ProverData,
		TxDataFromCalldata: t.TxDataFromCalldata,
		TxDataFromBlob:     t.TxDataFromBlob,
		BlobCommitments:    t.BlobCommitments,
		BlobProofs:         t.BlobProofs,
		BlobProofType:      t.BlobProofType,
	}
}

func (g *BatchGuestInput) UnmarshalJSON(data []byte) error {
	var dec batchGuestInputJSON
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}
	*g = *dec.GethType()
	return nil
}
