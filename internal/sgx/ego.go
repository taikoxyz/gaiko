package sgx

import (
	"crypto/ecdsa"
	"os"
	"path/filepath"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
)

type EgoProvider struct {
	args *flags.Arguments
}

var _ Provider = (*EgoProvider)(nil)

func NewEgoProvider(args *flags.Arguments) *EgoProvider {
	return &EgoProvider{
		args: args,
	}
}

func (p *EgoProvider) LoadQuote(key common.Address) ([]byte, error) {
	return getReport(key.Bytes())
}

func (p *EgoProvider) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	filename := filepath.Join(p.args.SecretDir, privKeyFilename)
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

func (p *EgoProvider) SavePrivateKey(privKey *ecdsa.PrivateKey) error {
	plainText := crypto.FromECDSA(privKey)
	// encrypt private key with a key derived from a measurement of the enclave.
	sealedText, err := ecrypto.SealWithUniqueKey(plainText, nil)
	if err != nil {
		return err
	}
	filename := filepath.Join(p.args.SecretDir, privKeyFilename)
	return os.WriteFile(filename, sealedText, 0600)
}

func (p *EgoProvider) SaveBootstrap(b *BootstrapData) error {
	filename := filepath.Join(p.args.ConfigDir, bootstrapInfoFilename)
	return b.SaveToFile(filename)
}

func getReport(userReport []byte) ([]byte, error) {
	report, err := enclave.GetRemoteReport(userReport)
	if err != nil {
		return nil, err
	}
	// NB: The first 16 bytes of the report are quote nonce.
	return report[16:], nil
}
