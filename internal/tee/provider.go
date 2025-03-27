package tee

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/taikoxyz/gaiko/internal/flags"
)

const (
	privKeyFilename       = "priv.gaiko.key"
	bootstrapInfoFilename = "bootstrap.gaiko.json"
)

// Provider is the interface that wraps the basic methods to interact with the TEE.
type Provider interface {
	// LoadQuote loads the quote from the TEE.
	LoadQuote(args *flags.Arguments, key common.Address) (Quote, error)
	// LoadPrivateKey loads the encrypted(mrenclave related) private key from the TEE.
	// The encrypted data only can be decrypted by the same instance(image).
	LoadPrivateKey(args *flags.Arguments) (*ecdsa.PrivateKey, error)
	// SavePrivateKey saves the encrypted(mrenclave related) private key to the TEE.
	SavePrivateKey(args *flags.Arguments, privKey *ecdsa.PrivateKey) error
	// SaveBootstrap saves the bootstrap data to the FS under the TEE(VM/Attach).
	SaveBootstrap(args *flags.Arguments, b *BootstrapData) error
}

// BootstrapData is the data structure representing the booting information.
type BootstrapData struct {
	PublicKey   hexutil.Bytes  `json:"public_key"`
	NewInstance common.Address `json:"new_instance"`
	Quote       hexutil.Bytes  `json:"quote"`
}

// SaveToFile saves the BootstrapData to a specified file in JSON format.
// The JSON output is indented for readability.
// Print details before save.
func (b *BootstrapData) SaveToFile(args *flags.Arguments) error {
	// 1. write to output
	stdioEncoder := json.NewEncoder(args.BootstrapWriter)
	if err := stdioEncoder.Encode(b); err != nil {
		return err
	}
	// 2. write to file
	filename := filepath.Join(args.ConfigDir, bootstrapInfoFilename)
	fmt.Printf("Bootstrap details saved in: %s\n", filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileEncoder := json.NewEncoder(file)
	fileEncoder.SetIndent("", "  ")
	return fileEncoder.Encode(b)
}
