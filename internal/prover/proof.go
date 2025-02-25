package prover

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
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
	return json.NewEncoder(os.Stdout).Encode(p)
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

func genSgxProof(
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
	pi, err := witness.NewPublicInput(driver, witness.SgxProofType, newInstance)
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

	proof := NewOneshotProof(args.InstanceID, newInstance, sign)
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
