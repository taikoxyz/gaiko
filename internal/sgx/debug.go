package sgx

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/internal/flags"
)

// DebugProvider purpose provider
type DebugProvider struct {
	privKey   *ecdsa.PrivateKey
	bootstrap *BootstrapData
}

var _ Provider = (*DebugProvider)(nil)

func NewDebugProvider(_ *flags.Arguments) *DebugProvider {
	return &DebugProvider{}
}

func (p *DebugProvider) LoadQuote(key common.Address) ([]byte, error) {
	var quote Quote
	return quote[:], nil
}

func (p *DebugProvider) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	return p.privKey, nil
}

func (p *DebugProvider) SavePrivateKey(privKey *ecdsa.PrivateKey) error {
	p.privKey = privKey
	return nil
}

func (p *DebugProvider) SaveBootstrap(b *BootstrapData) error {
	p.bootstrap = b
	fmt.Printf("Bootstrap details: %v\n", b)
	return nil
}
