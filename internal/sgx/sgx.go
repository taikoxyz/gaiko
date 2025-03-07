package sgx

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	privKeyFilename       = "priv.key"
	bootstrapInfoFilename = "bootstrap.json"
)

type Quote [432]byte

func BytesToQuote(b []byte) *Quote {
	var q Quote
	copy(q[:], b)
	return &q
}

func (q *Quote) Print() {
	fmt.Printf("Detected attestation type: enclave")
	fmt.Printf(
		"Extracted SGX quote with size = %d and the following fields:\n",
		len(q),
	)
	fmt.Printf(
		"  ATTRIBUTES.FLAGS: %x  [ Debug bit: %t ]\n",
		q[96:104],
		q[96]&2 > 0,
	)
	fmt.Printf("  ATTRIBUTES.XFRM:  %x\n", q[104:112])
	fmt.Printf("  MRENCLAVE:        %x\n", q[112:144])
	fmt.Printf("  MRSIGNER:         %x\n", q[176:208])
	fmt.Printf("  ISVPRODID:        %x\n", q[304:306])
	fmt.Printf("  ISVSVN:           %x\n", q[306:308])
	fmt.Printf("  REPORTDATA:       %x\n", q[368:400])
	fmt.Printf("                    %x\n", q[400:432])
}

type Provider interface {
	LoadQuote(key common.Address) ([]byte, error)
	LoadPrivateKey() (*ecdsa.PrivateKey, error)
	SavePrivateKey(privKey *ecdsa.PrivateKey) error
	SaveBootstrap(b *BootstrapData) error
}

type BootstrapData struct {
	PublicKey   hexutil.Bytes  `json:"public_key"`
	NewInstance common.Address `json:"new_instance"`
	Quote       hexutil.Bytes  `json:"quote"`
}

func (b *BootstrapData) SaveToFile(filename string) error {
	fmt.Printf("Bootstrap details saved in: %s\n", filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(b)
}
