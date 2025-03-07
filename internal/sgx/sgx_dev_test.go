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

func TestDevProvider(t *testing.T) {
	p := NewProvider(nil)
	q, err := p.LoadQuote(common.Address{})
	require.NoError(t, err)
	assert.Equal(t, devQuote, q)
	privKey, err := p.LoadPrivateKey()
	require.NoError(t, err)
	assert.Equal(t, devPrivKey, privKey)

	newInstance := crypto.PubkeyToAddress(devPrivKey.PublicKey)
	fmt.Printf("Instance address: %#x\n", newInstance)
	Quote(devQuote).Print()
}
