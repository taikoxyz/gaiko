package prover

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/tee"
	"github.com/taikoxyz/gaiko/internal/transition"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/taikoxyz/gaiko/pkg/keccak"
)

type ProofResponse struct {
	Proof           hexutil.Bytes  `json:"proof"`
	Quote           hexutil.Bytes  `json:"quote"`
	PublicKey       hexutil.Bytes  `json:"public_key"`
	InstanceAddress common.Address `json:"instance_address"`
	Input           common.Hash    `json:"input"`
}

func (p *ProofResponse) Output(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
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

func genAggregateProof(
	_ context.Context,
	args *flags.Arguments,
	sgxProvider tee.Provider,
) (*ProofResponse, error) {
	prevPrivKey, err := sgxProvider.LoadPrivateKey()
	if err != nil {
		return nil, err
	}
	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	var input witness.RawAggregationGuestInput
	err = json.NewDecoder(args.WitnessReader).Decode(&input)
	if err != nil {
		return nil, err
	}
	oldInstance := common.BytesToAddress(input.Proofs[0].Proof[4:24])
	curInstance := oldInstance
	for i, proof := range input.Proofs {
		pubKey, err := crypto.SigToPub(proof.Input.Bytes(), proof.Proof[24:])
		if err != nil {
			return nil, err
		}
		if crypto.PubkeyToAddress(*pubKey) != curInstance {
			return nil, fmt.Errorf("invalid proof[%d]", i)
		}
		curInstance = common.BytesToAddress(proof.Proof[4:24])
	}
	if newInstance != curInstance {
		return nil, fmt.Errorf("invalid instance: %#x", curInstance)
	}

	combinedHashes := make([]byte, 0, (len(input.Proofs)+2)*common.HashLength)
	combinedHashes = append(combinedHashes, addr2HashPadding[:]...)
	combinedHashes = append(combinedHashes, oldInstance.Bytes()...)
	combinedHashes = append(combinedHashes, addr2HashPadding[:]...)
	combinedHashes = append(combinedHashes, newInstance.Bytes()...)
	for _, proof := range input.Proofs {
		combinedHashes = append(combinedHashes, proof.Input.Bytes()...)
	}

	aggHash := keccak.Keccak(combinedHashes)
	sign, err := crypto.Sign(aggHash.Bytes(), prevPrivKey)
	if err != nil {
		return nil, err
	}

	proof := NewAggregateProof(args.SGXInstanceID, oldInstance, newInstance, sign)
	quote, err := sgxProvider.LoadQuote(newInstance)
	if err != nil {
		return nil, err
	}
	quote.Print()
	return &ProofResponse{
		Proof:           proof[:],
		Quote:           quote.Bytes(),
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           aggHash,
	}, nil
}

func genOneshotProof(
	ctx context.Context,
	args *flags.Arguments,
	wit witness.Witness,
	sgxProvider tee.Provider,
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
	quote.Print()
	return &ProofResponse{
		Proof:           proof[:],
		Quote:           quote.Bytes(),
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           piHash,
	}, nil
}
