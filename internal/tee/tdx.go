//go:build !dev

package tee

import (
	"crypto/ecdsa"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-tdx-guest/client"
	labi "github.com/google/go-tdx-guest/client/linuxabi"
	"github.com/taikoxyz/gaiko/internal/flags"
)

type TDXProvider struct {
}

var _ Provider = (*TDXProvider)(nil)

func NewTDXProvider() Provider {
	return &TDXProvider{}
}

func (p *TDXProvider) LoadQuote(args *flags.Arguments, key common.Address) ([]byte, error) {
	tdxQuoteProvider, err := client.GetQuoteProvider()
	if err != nil {
		return nil, err
	}

	var reportData64 [labi.TdReportDataSize]byte
	copy(reportData64[:], key.Bytes())

	return client.GetRawQuote(tdxQuoteProvider, reportData64)
}

func (p *TDXProvider) LoadPrivateKey(args *flags.Arguments) (*ecdsa.PrivateKey, error) {
	// TODO: encrypt private key with a key derived from a measurement of the enclave.
	panic("not implemented") // TODO: Implement
}

func (p *TDXProvider) SavePrivateKey(args *flags.Arguments, privKey *ecdsa.PrivateKey) error {
	// TODO: decrypt private key with a key derived from a measurement of the enclave.
	panic("not implemented") // TODO: Implement
}

func (p *TDXProvider) SaveBootstrap(args *flags.Arguments, b *BootstrapData) error {
	filename := filepath.Join(args.ConfigDir, bootstrapInfoFilename)
	return b.SaveToFile(filename)
}

func (p *TDXProvider) Quote(q []byte) Quote {
	return QuoteV4(q)
}
