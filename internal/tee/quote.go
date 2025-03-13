package tee

import (
	"fmt"
)

// Quote is the interface that wraps the basic methods to interact with the quote.
type Quote interface {
	// Bytes returns the origin data of the quote.
	Bytes() []byte
	// Print prints the quote in a human-readable format.
	Print()
}

type QuoteV3 []byte

func (q QuoteV3) Print() {
	fmt.Printf("Detected attestation type: enclave\n")
	if len(q) < 432 {
		fmt.Printf("Unexpected quote length: %d\n", len(q))
		return
	} else {
		fmt.Printf(
			"Extracted SGX quote with size = %d and the following fields:\n",
			len(q),
		)
	}
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

func (q QuoteV3) Bytes() []byte {
	return q
}

type QuoteV4 []byte

func (q QuoteV4) Print() {
	// TODO: parse and print quote v4
}

func (q QuoteV4) Bytes() []byte {
	return q
}
