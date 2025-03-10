//go:build dev

package tee

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevProvider(t *testing.T) {
	p := NewSGXProvider()
	q, err := p.LoadQuote(nil, common.Address{})
	require.NoError(t, err)
	assert.Equal(t, devQuoteV3, q.Bytes())
	privKey, err := p.LoadPrivateKey(nil)
	require.NoError(t, err)
	assert.Equal(t, devPrivKey, privKey)

	newInstance := crypto.PubkeyToAddress(devPrivKey.PublicKey)
	fmt.Printf("Instance address: %#x\n", newInstance)
	q.Print()
}
