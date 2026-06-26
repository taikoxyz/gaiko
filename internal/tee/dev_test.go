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
	p := NewSGXProvider(nil)
	privKey, err := p.LoadPrivateKey(nil)
	require.NoError(t, err)
	assert.Equal(t, devPrivKey, privKey)

	newInstance := crypto.PubkeyToAddress(devPrivKey.PublicKey)
	q, err := p.LoadQuote(nil, newInstance)
	require.NoError(t, err)
	assert.Len(t, q.Bytes(), sgxQuoteSize)
	assert.Equal(t, newInstance.Bytes(), q.Bytes()[sgxQuoteReportDataOffset:sgxQuoteReportDataOffset+common.AddressLength])

	fmt.Printf("Instance address: %#x\n", newInstance)
	q.Print()
}
