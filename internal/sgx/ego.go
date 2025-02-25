package sgx

import (
	"crypto/ecdsa"
	"os"
	"path"

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
	var extendedPubkey [64]byte
	copy(extendedPubkey[:], key.Bytes())
	return getReport(extendedPubkey[:])
}

func (p *EgoProvider) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	filename := path.Join(p.args.SecretDir, privKeyFilename)
	sealedText, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	plainText, err := ecrypto.Unseal(sealedText, nil)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(plainText)
}

func (p *EgoProvider) SavePrivateKey(privKey *ecdsa.PrivateKey) error {
	plainText := crypto.FromECDSA(privKey)
	sealedText, err := ecrypto.SealWithUniqueKey(plainText, nil)
	if err != nil {
		return err
	}
	filename := path.Join(p.args.SecretDir, privKeyFilename)
	return os.WriteFile(filename, sealedText, 0600)
}

func (p *EgoProvider) SaveBootstrap(b *BootstrapData) error {
	filename := path.Join(p.args.ConfigDir, bootstrapInfoFilename)
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
