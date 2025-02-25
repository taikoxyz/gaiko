package sgx

import (
	"crypto/ecdsa"
	"errors"
	"os"
	"path"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EgoProvider struct {
	secretDir  string
	userReport common.Address
}

var _ Provider = (*EgoProvider)(nil)

func NewEgoProvider(secretDir string) *EgoProvider {
	return &EgoProvider{
		secretDir: secretDir,
	}
}

func (p *EgoProvider) LoadQuote() ([]byte, error) {
	if len(p.userReport) == 0 {
		return nil, errors.New("user report is not set")
	}
	var extendedPubkey [64]byte
	copy(extendedPubkey[:], p.userReport.Bytes())
	return getReport(extendedPubkey[:])
}

func (p *EgoProvider) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	filename := path.Join(p.secretDir, privKeyFilename)
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
	filename := path.Join(p.secretDir, privKeyFilename)
	return os.WriteFile(filename, sealedText, 0600)
}

func (p *EgoProvider) SavePublicKey(key common.Address) error {
	p.userReport = key
	return nil
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
