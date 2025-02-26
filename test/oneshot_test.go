package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/taikoxyz/gaiko/test/fixtures"
)

func TestOneshot(t *testing.T) {
	inputs, err := fixtures.GetSingleInputs()
	require.NoError(t, err)

	for _, input := range inputs {
		t.Run(fmt.Sprintf("block: %d", input.Block.NumberU64()), func(t *testing.T) {
			args := &flags.Arguments{
				SecretDir:  "",
				ConfigDir:  "",
				SgxType:    "debug",
				InstanceID: 0,
			}
			sgxProver := prover.NewSGXProver(args)
			proof, err := sgxProver.Oneshot(context.Background())
			require.NoError(t, err)
			t.Logf("proof: %s", proof)
		})
	}
}
