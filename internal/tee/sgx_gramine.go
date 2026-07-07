package tee

import (
	"crypto/ecdsa"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
)

const (
	attestationQuoteDeviceFile = "/dev/attestation/quote"
	// attestationTypeDeviceFile           = "/dev/attestation/attestation_type"
	attestationUserReportDataDeviceFile = "/dev/attestation/user_report_data"
)

type SGXGramineProvider struct{}

var _ Provider = (*SGXGramineProvider)(nil)

func NewGramineProvider() *SGXGramineProvider {
	return &SGXGramineProvider{}
}

func (p *SGXGramineProvider) LoadQuote(args *flags.Arguments, key common.Address) (Quote, error) {
	q, err := getQuote(key)
	if err != nil {
		return nil, err
	}
	return QuoteV3(q), nil
}

func (p *SGXGramineProvider) LoadPrivateKey(args *flags.Arguments) (*ecdsa.PrivateKey, error) {
	return loadPrivKey(args.SecretDir)
}

func (p *SGXGramineProvider) SavePrivateKey(
	args *flags.Arguments,
	privKey *ecdsa.PrivateKey,
) error {
	return savePrivKey(args.SecretDir, privKey)
}

func (p *SGXGramineProvider) SaveBootstrap(args *flags.Arguments, b *BootstrapData) error {
	return b.SaveToFile(args)
}

func saveAttestationUserReportData(pubKey common.Address) error {
	extendedPubkey := make([]byte, 64)
	copy(extendedPubkey, pubKey.Bytes())
	userReportDataFile, err := os.OpenFile(attestationUserReportDataDeviceFile, os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	defer userReportDataFile.Close()
	if _, err := userReportDataFile.Write(extendedPubkey); err != nil {
		return err
	}
	return nil
}

func getQuote(pubkey common.Address) ([]byte, error) {
	err := saveAttestationUserReportData(pubkey)
	if err != nil {
		return nil, err
	}
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
	ownerRead := mode&0o400 != 0  // 0400 is read permission for owner (r--------)
	ownerWrite := mode&0o200 != 0 // 0200 is write permission for owner (-w-------)
	// Read-only means readable but not writable
	return ownerRead && !ownerWrite
}

func isFile(fileInfo os.FileInfo) bool {
	return fileInfo.Mode().IsRegular()
}

func loadPrivKey(secretsDir string) (*ecdsa.PrivateKey, error) {
	privKeyPath := filepath.Join(secretsDir, privKeyFilename)
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

func savePrivKey(secretDir string, privKey *ecdsa.PrivateKey) error {
	privKeyPath := filepath.Join(secretDir, privKeyFilename)
	return crypto.SaveECDSA(privKeyPath, privKey)
}
