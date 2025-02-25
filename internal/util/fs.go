package util

import (
	"crypto/ecdsa"
	"errors"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/crypto"
)

const privKeyFilename = "priv.key"

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

func LoadPrivKey(secretsDir string) (*ecdsa.PrivateKey, error) {
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
