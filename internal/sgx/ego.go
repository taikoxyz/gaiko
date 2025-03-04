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
	var extendedPubKey [64]byte
	copy(extendedPubKey[:], key.Bytes())
	return getReport(extendedPubKey[:])
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
	// get empty local report to use it as target info
	report, err := enclave.GetLocalReport(nil, nil)
	if err != nil {
		return nil, err
	}
	// get report for target
	return enclave.GetLocalReport(userReport, report)
}
