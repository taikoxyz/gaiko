package prover

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/sgx"
	"github.com/taikoxyz/gaiko/internal/transition"
	"github.com/taikoxyz/gaiko/internal/witness"
)

type ProofResponse struct {
	Proof           hexutil.Bytes  `json:"proof"`
	Quote           hexutil.Bytes  `json:"quote"`
	PublicKey       hexutil.Bytes  `json:"public_key"`
	InstanceAddress common.Address `json:"instance_address"`
	Input           common.Hash    `json:"input"`
}

func (p *ProofResponse) Stdout() error {
	err := json.NewEncoder(os.Stdout).Encode(p)
	if err != nil {
		return err
	}
	quote := p.Quote

	fmt.Printf("Detected attestation type: enclave")
	fmt.Printf(
		"Extracted SGX quote with size = %d and the following fields:\n",
		len(quote),
	)
	fmt.Printf(
		"  ATTRIBUTES.FLAGS: %s  [ Debug bit: %t ]\n",
		hex.EncodeToString(quote[96:104]),
		quote[96]&2 > 0,
	)
	fmt.Printf("  ATTRIBUTES.XFRM:  %s\n", hex.EncodeToString(quote[104:112]))
	fmt.Printf("  MRENCLAVE:        %s\n", hex.EncodeToString(quote[112:144]))
	fmt.Printf("  MRSIGNER:         %s\n", hex.EncodeToString(quote[176:208]))
	fmt.Printf("  ISVPRODID:        %s\n", hex.EncodeToString(quote[304:306]))
	fmt.Printf("  ISVSVN:           %s\n", hex.EncodeToString(quote[306:308]))
	fmt.Printf("  REPORTDATA:       %s\n", hex.EncodeToString(quote[368:400]))
	fmt.Printf("                    %s\n", hex.EncodeToString(quote[400:432]))
	return nil
}

type OneshotProof [89]byte

func (p *OneshotProof) Hex() string {
	return hex.EncodeToString(p[:])
}

func NewOneshotProof(instanceID uint32, newInstance common.Address, sign []byte) *OneshotProof {
	var proof OneshotProof
	binary.BigEndian.PutUint32(proof[:4], instanceID)
	copy(proof[4:24], newInstance.Bytes())
	copy(proof[24:], sign)
	return &proof
}

type AggregateProof [109]byte

func (p *AggregateProof) Hex() string {
	return hex.EncodeToString(p[:])
}

func NewAggregateProof(
	instanceID uint32,
	oldInstance, newInstance common.Address,
	sign []byte,
) *AggregateProof {
	var proof AggregateProof
	binary.BigEndian.PutUint32(proof[:4], instanceID)
	copy(proof[4:24], oldInstance.Bytes())
	copy(proof[24:44], newInstance.Bytes())
	copy(proof[44:], sign)
	return &proof
}

func genOneshotProof(
	ctx context.Context,
	args *flags.Arguments,
	driver witness.GuestDriver,
	provider sgx.Provider,
) (*ProofResponse, error) {
	err := json.NewDecoder(os.Stdin).Decode(driver)
	if err != nil {
		return nil, err
	}
	err = transition.ExecuteGuestDriver(ctx, args, driver)
	if err != nil {
		return nil, err
	}
	prevPrivKey, err := provider.LoadPrivateKey()
	if err != nil {
		return nil, err
	}

	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	pi, err := witness.NewPublicInput(driver, witness.SGXProofType, newInstance)
	if err != nil {
		return nil, err
	}
	piHash, err := pi.Hash()
	if err != nil {
		return nil, err
	}

	sign, err := crypto.Sign(piHash.Bytes(), prevPrivKey)
	if err != nil {
		return nil, err
	}

	proof := NewOneshotProof(args.SGXInstanceID, newInstance, sign)
	quote, err := provider.LoadQuote(newInstance)
	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		Proof:           proof[:],
		Quote:           quote,
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           piHash,
	}, nil
}
