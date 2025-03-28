package prover

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytesUnmarshal(t *testing.T) {
	raw := `{"tx_data": [1,2,3,4,255,255,255,255,255]}`
	type Foo struct {
		TxData []byte `json:"tx_data"`
	}
	var foo Foo
	err := json.Unmarshal([]byte(raw), &foo)
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3, 4, 255, 255, 255, 255, 255}, foo.TxData)
}
