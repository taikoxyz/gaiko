package witness

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/log"
	gaikoTypes "github.com/taikoxyz/gaiko/internal/types"
)

type batchGuestInputJSON struct {
	Inputs []*singleGuestInputJSON   `json:"inputs"`
	Taiko  *taikoGuestBatchInputJSON `json:"taiko"`
}

func (g *batchGuestInputJSON) GethType() *BatchGuestInput {
	if g == nil {
		log.Warn("missing batchGuestInputJSON when converting to GethType")
		return nil
	}
	guestInputs := make([]*SingleGuestInput, len(g.Inputs))
	res := &BatchGuestInput{
		Inputs: guestInputs,
		Taiko:  g.Taiko.GethType(),
	}
	for i, input := range g.Inputs {
		input := input.GethType()
		if input != nil {
			input.parent = res
		}
		guestInputs[i] = input
	}
	return res
}

type taikoGuestDataSourceJSON struct {
	TxDataFromCalldata []byte                  `json:"tx_data_from_calldata"`
	TxDataFromBlob     [][eth.BlobSize]byte    `json:"tx_data_from_blob"`
	BlobCommitments    *[][commitmentSize]byte `json:"blob_commitments"`
	BlobProofs         *[][proofSize]byte      `json:"blob_proofs"`
	BlobProofType      BlobProofType           `json:"blob_proof_type"`
}

func (t *taikoGuestDataSourceJSON) GethType() *TaikoGuestDataSource {
	if t == nil {
		log.Warn("missing taikoGuestDataSourceJSON when converting to GethType")
		return nil
	}
	return &TaikoGuestDataSource{
		TxDataFromCalldata: t.TxDataFromCalldata,
		TxDataFromBlob:     t.TxDataFromBlob,
		BlobCommitments:    t.BlobCommitments,
		BlobProofs:         t.BlobProofs,
		BlobProofType:      t.BlobProofType,
	}
}

type taikoGuestBatchInputJSON struct {
	BatchID       uint64                      `json:"batch_id"`
	L1Header      *gaikoTypes.Header          `json:"l1_header"`
	BatchProposed *blockProposedJSON          `json:"batch_proposed"`
	ChainSpec     *ChainSpec                  `json:"chain_spec"`
	ProverData    *TaikoProverData            `json:"prover_data"`
	DataSources   []*taikoGuestDataSourceJSON `json:"data_sources"`
}

func (t *taikoGuestBatchInputJSON) GethType() *TaikoGuestBatchInput {
	dataSources := make([]*TaikoGuestDataSource, 0, len(t.DataSources))
	for _, ds := range t.DataSources {
		dataSources = append(dataSources, ds.GethType())
	}
	return &TaikoGuestBatchInput{
		BatchID:       t.BatchID,
		L1Header:      t.L1Header.GethType(),
		BatchProposed: t.BatchProposed.GethType(),
		ChainSpec:     t.ChainSpec,
		ProverData:    t.ProverData,
		DataSources:   dataSources,
	}
}

func (g *BatchGuestInput) UnmarshalJSON(data []byte) error {
	// Handle backward compatibility: old format has tx_data_from_calldata directly in taiko,
	// new format has it in data_sources array
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if taikoRaw, exists := raw["taiko"]; exists {
		var taikoObj map[string]json.RawMessage
		if err := json.Unmarshal(taikoRaw, &taikoObj); err == nil {
			// Check if this is old format (has tx_data_from_calldata but no data_sources)
			_, hasOldFormat := taikoObj["tx_data_from_calldata"]
			_, hasNewFormat := taikoObj["data_sources"]

			if hasOldFormat && !hasNewFormat {
				// Convert old format to new format by wrapping in data_sources array
				dataSource := make(map[string]json.RawMessage)
				for _, key := range []string{"tx_data_from_calldata", "tx_data_from_blob", "blob_commitments", "blob_proofs", "blob_proof_type"} {
					if val, ok := taikoObj[key]; ok {
						dataSource[key] = val
						delete(taikoObj, key)
					}
				}
				dataSourcesArray, _ := json.Marshal([]map[string]json.RawMessage{dataSource})
				taikoObj["data_sources"] = dataSourcesArray

				// Reconstruct taiko object
				newTaikoRaw, _ := json.Marshal(taikoObj)
				raw["taiko"] = newTaikoRaw
				data, _ = json.Marshal(raw)
			}
		}
	}

	var dec batchGuestInputJSON
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}
	*g = *dec.GethType()
	return nil
}
