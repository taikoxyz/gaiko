//go:build dev

package sgx

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestProvider(t *testing.T) {
	p := NewProvider(nil)
	q, err := p.LoadQuote(common.Address{})
	require.NoError(t, err)
	assert.Equal(t, testQuote, q)
	privKey, err := p.LoadPrivateKey()
	require.NoError(t, err)
	assert.Equal(t, testPrivKey, privKey)

	newInstance := crypto.PubkeyToAddress(testPrivKey.PublicKey)
	fmt.Printf("Instance address: %#x\n", newInstance)
}
