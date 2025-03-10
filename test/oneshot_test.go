package test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/taikoxyz/gaiko/test/fixtures"
)

func TestOneshot(t *testing.T) {
	inputs, err := fixtures.GetSingleInputs()
	require.NoError(t, err)

	for idx, input := range inputs {
		t.Run(fmt.Sprintf("task: %d", idx), func(t *testing.T) {
			args := &flags.Arguments{
				SecretDir:     "",
				ConfigDir:     "",
				SGXType:       "debug",
				SGXInstanceID: 0,
				WitnessReader: bytes.NewBuffer(input),
				ProofWriter:   os.Stdout,
			}
			sgxProver := prover.NewSGXProver(args)
			err := sgxProver.Oneshot(context.Background(), args)
			require.NoError(t, err)
		})
	}
}
