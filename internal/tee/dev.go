//go:build dev

package tee

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
)

const (
	devPrivateKeyEnv     = "GAIKO_DEV_PRIVATE_KEY"
	devPrivateKeyFileEnv = "GAIKO_DEV_PRIVATE_KEY_FILE"

	sgxQuoteSize             = 432
	sgxQuoteMrEnclaveOffset  = 112
	sgxQuoteMrSignerOffset   = 176
	sgxQuoteReportDataOffset = 368
	sgxQuoteReportDataSize   = 64

	// Fixed mock measurements keep dev output recognizable without embedding
	// real signing material.
	devMrEnclave = "7cb1afb4d3505b9028d9aec761be3541a703b072eee5800be2f98e844f1cebcc"
	devMrSigner  = "97f37974b1a9a1f64b2e50b820a79721078df06e1268a303bd8427100d587f44"
)

var (
	devQuoteV3 []byte
	devPrivKey *ecdsa.PrivateKey
)

func init() {
	var err error
	devPrivKey, err = loadDevPrivateKey()
	if err != nil {
		panic(err)
	}
	devQuoteV3 = newDevQuoteV3(common.Address{})
}

func loadDevPrivateKey() (*ecdsa.PrivateKey, error) {
	rawPrivKey := strings.TrimSpace(os.Getenv(devPrivateKeyEnv))
	if rawPrivKey == "" {
		path := strings.TrimSpace(os.Getenv(devPrivateKeyFileEnv))
		if path != "" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", devPrivateKeyFileEnv, err)
			}
			rawPrivKey = strings.TrimSpace(string(data))
		}
	}
	if rawPrivKey != "" {
		return crypto.HexToECDSA(strings.TrimPrefix(rawPrivKey, "0x"))
	}
	return crypto.GenerateKey()
}

func newDevQuoteV3(key common.Address) []byte {
	q := make([]byte, sgxQuoteSize)
	copy(q[sgxQuoteMrEnclaveOffset:sgxQuoteMrEnclaveOffset+common.HashLength], common.FromHex(devMrEnclave))
	copy(q[sgxQuoteMrSignerOffset:sgxQuoteMrSignerOffset+common.HashLength], common.FromHex(devMrSigner))
	setQuoteReportData(q, key)
	return q
}

func setQuoteReportData(q []byte, key common.Address) {
	reportData := q[sgxQuoteReportDataOffset : sgxQuoteReportDataOffset+sgxQuoteReportDataSize]
	clear(reportData)
	copy(reportData, key.Bytes())
}

type quoteVersion int

const (
	quoteV3Version quoteVersion = 3
	quoteV4Version quoteVersion = 4
)

type DevProvider struct {
	quoteVersion quoteVersion
}

func NewSGXProvider(_ *flags.Arguments) Provider {
	return &DevProvider{
		quoteVersion: quoteV3Version,
	}
}

func NewTDXProvider(_ *flags.Arguments) Provider {
	return &DevProvider{
		quoteVersion: quoteV4Version,
	}
}

func (p *DevProvider) LoadQuote(args *flags.Arguments, key common.Address) (Quote, error) {
	return QuoteV3(newDevQuoteV3(key)), nil
}

func (p *DevProvider) LoadPrivateKey(args *flags.Arguments) (*ecdsa.PrivateKey, error) {
	return devPrivKey, nil
}

func (p *DevProvider) SavePrivateKey(args *flags.Arguments, privKey *ecdsa.PrivateKey) error {
	// rewrite the private key with the mock one
	*privKey = *devPrivKey
	return nil
}

func (p *DevProvider) SaveBootstrap(args *flags.Arguments, b *BootstrapData) error {
	return nil
}
