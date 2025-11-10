package tests

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

type SingleGuestOutput struct {
	Hash common.Hash `json:"hash"`
}

func TestSingle(t *testing.T) {
	inputs, err := fixtures.GetSingleInputs()
	require.NoError(t, err)

	for id, input := range inputs {
		t.Run(fmt.Sprintf("task:%d", id), func(t *testing.T) {
			var output prover.ProofResponse
			var b bytes.Buffer
			args := &flags.Arguments{
				SecretDir:      "",
				ConfigDir:      "",
				SGXType:        "debug",
				ProofType:      witness.NativeProofType,
				SGXInstanceID:  0,
				SGXInstanceIDs: make(map[string]uint32),
				WitnessReader:  bytes.NewBuffer(input.Input),
				ProofWriter:    &b,
			}
			sgxProver := prover.NewSGXProver(args)
			err = sgxProver.Oneshot(context.Background(), args)
			require.NoError(t, err)

			err := json.NewDecoder(&b).Decode(&output)
			require.NoError(t, err)

			// Verify the output
			var expectedOutput SingleGuestOutput
			err = json.Unmarshal(input.Output, &expectedOutput)
			require.NoError(t, err)

			assert.Equal(t, expectedOutput.Hash, output.Input)
		})
	}
}
