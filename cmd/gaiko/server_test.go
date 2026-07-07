package main

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type raikoLikeRemoteSgxResponse struct {
	Status      string                `json:"status"`
	Message     string                `json:"message"`
	SGXResponse raikoLikeSgxResponse  `json:"proof"`
}

type raikoLikeSgxResponse struct {
	Proof string      `json:"proof"`
	Quote string      `json:"quote"`
	Input common.Hash `json:"input"`
}

func TestErrorResponseProofPayloadIsRaikoCompatible(t *testing.T) {
	response := Response{
		Status:  "error",
		Message: "boom",
		Proof:   emptyProofResponsePayload,
	}

	raw, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded raikoLikeRemoteSgxResponse
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, "error", decoded.Status)
	require.Equal(t, "boom", decoded.Message)
	require.Equal(t, "0x", decoded.SGXResponse.Proof)
	require.Equal(t, "0x", decoded.SGXResponse.Quote)
	require.Equal(t, common.Hash{}, decoded.SGXResponse.Input)
}
