package sgx

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

const privKeyFilename = "priv.key"

type Provider interface {
	LoadQuote() ([]byte, error)
	LoadPrivateKey() (*ecdsa.PrivateKey, error)
	SavePrivateKey(privKey *ecdsa.PrivateKey) error
	SavePublicKey(key common.Address) error
}
