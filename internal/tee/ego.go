package tee

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
)

type EgoProvider struct {
}

var _ Provider = (*EgoProvider)(nil)

func NewEgoProvider() *EgoProvider {
	return &EgoProvider{}
}

func (p *EgoProvider) LoadQuote(args *flags.Arguments, key common.Address) (Quote, error) {
	q, err := getRemoteReport(key.Bytes())
	if err != nil {
		return nil, err
	}
	return QuoteV3(q), nil
}

func (p *EgoProvider) LoadPrivateKey(args *flags.Arguments) (*ecdsa.PrivateKey, error) {
	filename := filepath.Join(args.SecretDir, privKeyFilename)
	sealedText, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// decrypt private key with a key derived from a measurement of the enclave.
	plainText, err := ecrypto.Unseal(sealedText, nil)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(plainText)
}

func (p *EgoProvider) SavePrivateKey(args *flags.Arguments, privKey *ecdsa.PrivateKey) error {
	plainText := crypto.FromECDSA(privKey)
	// encrypt private key with a key derived from a measurement of the enclave.
	sealedText, err := ecrypto.SealWithUniqueKey(plainText, nil)
	if err != nil {
		return err
	}
	filename := filepath.Join(args.SecretDir, privKeyFilename)
	return os.WriteFile(filename, sealedText, 0600)
}

func (p *EgoProvider) SaveBootstrap(args *flags.Arguments, b *BootstrapData) error {
	filename := filepath.Join(args.ConfigDir, bootstrapInfoFilename)
	return b.SaveToFile(filename)
}

func getRemoteReport(userReport []byte) ([]byte, error) {
	report, err := enclave.GetRemoteReport(userReport)
	if err != nil {
		return nil, err
	}
	// NB: The first 16 bytes of the report are quote nonce.
	if len(report) < 16 {
		return nil, fmt.Errorf("unexpected report length: %d", len(report))
	}
	return report[16:], nil
}
