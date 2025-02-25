package sgx

import (
	"crypto/ecdsa"
	"errors"
	"io"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	attestationQuoteDeviceFile          = "/dev/attestation/quote"
	attestationTypeDeviceFile           = "/dev/attestation/attestation_type"
	attestationUserReportDataDeviceFile = "/dev/attestation/user_report_data"
	bootstrapInfoFilename               = "bootstrap.json"
)

type GramineProvider struct {
	secretDir string
}

var _ Provider = (*GramineProvider)(nil)

func NewGramineProvider(secretDir string) *GramineProvider {
	return &GramineProvider{
		secretDir: secretDir,
	}
}

func (p *GramineProvider) LoadQuote() ([]byte, error) {
	return getSgxQuote()
}

func (p *GramineProvider) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	return loadPrivKey(p.secretDir)
}

func (p *GramineProvider) SavePrivateKey(privKey *ecdsa.PrivateKey) error {
	return SavePrivKey(p.secretDir, privKey)
}

func (p *GramineProvider) SavePublicKey(key common.Address) error {
	saveAttestationUserReportData(key)
	return nil
}

func saveAttestationUserReportData(pubkey common.Address) error {
	extendedPubkey := make([]byte, 64)
	copy(extendedPubkey, pubkey.Bytes())
	userReportDataFile, err := os.OpenFile(attestationUserReportDataDeviceFile, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer userReportDataFile.Close()
	if _, err := userReportDataFile.Write(extendedPubkey); err != nil {
		return err
	}
	return nil
}

func getSgxQuote() ([]byte, error) {
	quoteFile, err := os.Open(attestationQuoteDeviceFile)
	if err != nil {
		return nil, err
	}
	defer quoteFile.Close()
	quote, err := io.ReadAll(quoteFile)
	if err != nil {
		return nil, err
	}
	return quote, nil
}

func isReadOnly(fileInfo os.FileInfo) bool {
	// Get file mode (permissions)
	mode := fileInfo.Mode()
	// Check owner permissions (bits 8-6: rwx)
	ownerRead := mode&0400 != 0  // 0400 is read permission for owner (r--------)
	ownerWrite := mode&0200 != 0 // 0200 is write permission for owner (-w-------)
	// Read-only means readable but not writable
	return ownerRead && !ownerWrite
}

func isFile(fileInfo os.FileInfo) bool {
	return fileInfo.Mode().IsRegular()
}

func loadPrivKey(secretsDir string) (*ecdsa.PrivateKey, error) {
	privKeyPath := path.Join(secretsDir, privKeyFilename)
	fileInfo, err := os.Stat(privKeyPath)
	if err != nil {
		return nil, err
	}
	if isFile(fileInfo) {
		// only readonly private key file was allowed
		if isReadOnly(fileInfo) {
			return crypto.LoadECDSA(privKeyPath)
		} else {
			return nil, errors.New("file exists but has wrong permissions")
		}
	}
	return nil, errors.New("file does not exist")
}

func SavePrivKey(secretDir string, privKey *ecdsa.PrivateKey) error {
	privKeyPath := path.Join(secretDir, privKeyFilename)
	return crypto.SaveECDSA(privKeyPath, privKey)
}
