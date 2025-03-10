package prover

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"

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

func (p *ProofResponse) Output(w io.Writer) error {
	err := json.NewEncoder(w).Encode(p)
	if err != nil {
		return err
	}
	sgx.Quote(p.Quote).Print()
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
	wit witness.Witness,
	sgxProvider sgx.Provider,
) (*ProofResponse, error) {
	err := json.NewDecoder(args.WitnessReader).Decode(wit)
	if err != nil {
		return nil, err
	}
	err = transition.Execute(ctx, args, wit)
	if err != nil {
		return nil, err
	}
	prevPrivKey, err := sgxProvider.LoadPrivateKey()
	if err != nil {
		return nil, err
	}

	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	pi, err := witness.NewPublicInput(wit, witness.PivotProofType, newInstance)
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
	quote, err := sgxProvider.LoadQuote(newInstance)
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
