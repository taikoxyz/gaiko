package sgx

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/internal/flags"
)

const privKeyFilename = "priv.key"

type Provider interface {
	LoadQuote(key common.Address) ([]byte, error)
	LoadPrivateKey() (*ecdsa.PrivateKey, error)
	SavePrivateKey(privKey *ecdsa.PrivateKey) error
}

func NewProvider(typ, secretDir string) Provider {
	if typ == flags.GramineSGXType {
		return NewGramineProvider(secretDir)
	}
	return NewEgoProvider(secretDir)
}
