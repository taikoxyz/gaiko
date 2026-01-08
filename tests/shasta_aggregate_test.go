package tests

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/taikoxyz/gaiko/tests/fixtures"
)

func TestShastaAggregationHash(t *testing.T) {
	inputBytes, err := fixtures.ReadShastaFixture("agg-input-52.json")
	require.NoError(t, err)
	outputBytes, err := fixtures.ReadShastaFixture("agg-output-52.json")
	require.NoError(t, err)

	var input witness.ShastaRawAggregationGuestInput
	require.NoError(t, json.Unmarshal(inputBytes, &input))

	var expected BatchGuestOutput
	require.NoError(t, json.Unmarshal(outputBytes, &expected))

	actual, err := witness.ShastaPCDAggregationHash(input.ProofCarryDataVec, common.Address{})
	require.NoError(t, err)
	require.Equal(t, expected.Hash, actual)
}

