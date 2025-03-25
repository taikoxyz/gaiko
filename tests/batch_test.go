package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/taikoxyz/gaiko/tests/fixtures"
)

type BatchGuestOutput struct {
	Hash common.Hash `json:"hash"`
}

func TestBatch(t *testing.T) {
	inputs, err := fixtures.GetBatchInputs()
	require.NoError(t, err)

	for id, input := range inputs {
		t.Run(fmt.Sprintf("task: %d", id), func(t *testing.T) {
			var output prover.ProofResponse
			var b bytes.Buffer
			args := &flags.Arguments{
				SecretDir:     "",
				ConfigDir:     "",
				SGXType:       "debug",
				ProverType:    witness.PivotProofType,
				SGXInstanceID: 0,
				WitnessReader: bytes.NewBuffer(input.Input),
				ProofWriter:   &b,
			}
			sgxProver := prover.NewSGXProver(args)
			err = sgxProver.BatchOneshot(context.Background(), args)
			require.NoError(t, err)

			err := json.NewDecoder(&b).Decode(&output)
			require.NoError(t, err)

			// Verify the output
			var expectedOutput BatchGuestOutput
			err = json.Unmarshal(input.Output, &expectedOutput)
			require.NoError(t, err)

			assert.Equal(t, expectedOutput.Hash, output.Input)
		})
	}
}
