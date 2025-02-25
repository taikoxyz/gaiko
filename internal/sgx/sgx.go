package sgx

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/internal/flags"
)

const privKeyFilename = "priv.key"

type Provider interface {
	LoadQuote() ([]byte, error)
	LoadPrivateKey() (*ecdsa.PrivateKey, error)
	SavePrivateKey(privKey *ecdsa.PrivateKey) error
	SavePublicKey(key common.Address) error
}

func NewProvider(typ, secretDir string) Provider {
	if typ == flags.GramineSGXType {
		return NewGramineProvider(secretDir)
	}
	return NewEgoProvider(secretDir)
}
