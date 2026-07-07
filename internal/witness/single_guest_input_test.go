package witness

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestTaikoProverDataUnmarshalDesignatedProverNull(t *testing.T) {
	payload := []byte(`{
		"prover":"0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc",
		"designated_prover": null,
		"graffiti":"0x0000000000000000000000000000000000000000000000000000000000000000"
	}`)

	var data TaikoProverData
	require.NoError(t, json.Unmarshal(payload, &data))
	require.Equal(t, common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc"), data.ActualProver)
	require.Equal(t, (common.Address{}), data.DesignatedProver)
	require.Equal(t, common.Hash{}, data.Graffiti)
}

func TestTaikoProverDataUnmarshalActualProverFallback(t *testing.T) {
	payload := []byte(`{
		"actual_prover":"0x0000000000000000000000000000000000000001",
		"graffiti":"0x0000000000000000000000000000000000000000000000000000000000000000"
	}`)

	var data TaikoProverData
	require.NoError(t, json.Unmarshal(payload, &data))
	require.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000001"), data.ActualProver)
	require.Equal(t, (common.Address{}), data.DesignatedProver)
}
