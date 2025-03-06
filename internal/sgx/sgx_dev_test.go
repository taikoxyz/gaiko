//go:build dev

package sgx

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestProvider(t *testing.T) {
	p := NewProvider(nil)
	q, err := p.LoadQuote(common.Address{})
	require.NoError(t, err)
	assert.Equal(t, testQuote, q)
}
