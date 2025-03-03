package sgx

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/taikoxyz/gaiko/internal/flags"
)

const (
	privKeyFilename       = "priv.key"
	bootstrapInfoFilename = "bootstrap.json"
)

type Quote [432]byte

type Provider interface {
	LoadQuote(key common.Address) ([]byte, error)
	LoadPrivateKey() (*ecdsa.PrivateKey, error)
	SavePrivateKey(privKey *ecdsa.PrivateKey) error
	SaveBootstrap(b *BootstrapData) error
}

func NewProvider(args *flags.Arguments) Provider {
	switch args.SGXType {
	case flags.DebugSGXType:
		return NewDebugProvider(args)
	case flags.GramineSGXType:
		return NewGramineProvider(args)
	default:
		return NewEgoProvider(args)
	}
}

type BootstrapData struct {
	PublicKey   hexutil.Bytes  `json:"public_key"`
	NewInstance common.Address `json:"new_instance"`
	Quote       hexutil.Bytes  `json:"quote"`
}

func (b *BootstrapData) SaveToFile(filename string) error {
	fmt.Printf("Bootstrap details saved in: %s \n", filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(b)
}
