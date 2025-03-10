package tee

import (
	"fmt"
)

const (
	privKeyFilename       = "priv.key"
	bootstrapInfoFilename = "bootstrap.json"
)

type QuoteV3 []byte

func (q QuoteV3) Print() {
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
