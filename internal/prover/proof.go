package prover

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
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

func NewDefaultProofResponse() ProofResponse {
	return ProofResponse{
		Proof:           hexutil.MustDecode("0xdefac0de"),
		Quote:           hexutil.MustDecode("0xdefac0de"),
		PublicKey:       hexutil.MustDecode("0xdefac0de"),
		InstanceAddress: common.Address{},
		Input:           common.Hash{},
	}
}

func (p *ProofResponse) Output(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}

func NewOneshotProof(instanceID uint32, newInstance common.Address, sign []byte) []byte {
	var proof [89]byte
	binary.BigEndian.PutUint32(proof[:4], instanceID)
	copy(proof[4:24], newInstance.Bytes())
	copy(proof[24:], sign)
	return proof[:]
}

func NewAggregateProof(
	instanceID uint32,
	oldInstance, newInstance common.Address,
	sign []byte,
) []byte {
	var proof [109]byte
	binary.BigEndian.PutUint32(proof[:4], instanceID)
	copy(proof[4:24], oldInstance.Bytes())
	copy(proof[24:44], newInstance.Bytes())
	copy(proof[44:], sign)
	return proof[:]
}

func genAggregateProof(
	_ context.Context,
	args *flags.Arguments,
	provider tee.Provider,
) error {
	args.UpdateSGXInstanceID(witness.PacayaHardFork)
	prevPrivKey, err := provider.LoadPrivateKey(args)
	if err != nil {
		return err
	}
	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	var input witness.RawAggregationGuestInput
	err = json.NewDecoder(args.WitnessReader).Decode(&input)
	if err != nil {
		return err
	}
	log.Info("receive input: ", "input", input)
	oldInstance := common.BytesToAddress(input.Proofs[0].Proof[4:24])
	curInstance := oldInstance
	for i, proof := range input.Proofs {
		pubKey, err := SigToPub(proof.Input.Bytes(), proof.Proof[24:])
		if err != nil {
			return err
		}
		if crypto.PubkeyToAddress(*pubKey) != curInstance {
			return fmt.Errorf("invalid proof[%d]", i)
		}
		curInstance = common.BytesToAddress(proof.Proof[4:24])
	}
	if newInstance != curInstance {
		return fmt.Errorf("invalid instance: %#x", curInstance)
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
	sign, err := Sign(aggHash.Bytes(), prevPrivKey)
	if err != nil {
		return err
	}

	proof := NewAggregateProof(args.SGXInstanceID, oldInstance, newInstance, sign)
	quote, err := provider.LoadQuote(args, newInstance)
	if err != nil {
		return err
	}
	quote.Print()
	return (&ProofResponse{
		Proof:           proof,
		Quote:           quote.Bytes(),
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           aggHash,
	}).Output(args.ProofWriter)
}

func genOneshotProof(
	ctx context.Context,
	args *flags.Arguments,
	guestInput witness.GuestInput,
	provider tee.Provider,
) error {
	err := json.NewDecoder(args.WitnessReader).Decode(guestInput)
	if err != nil {
		return err
	}
	args.UpdateSGXInstanceID(guestInput.BlockProposed().HardFork())
	log.Info("Start generate proof: ", "id", guestInput.ID())
	err = transition.ExecuteAndVerify(ctx, args, guestInput)
	if err != nil {
		return err
	}
	prevPrivKey, err := provider.LoadPrivateKey(args)
	if err != nil {
		return err
	}

	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	if args.SGXType == "debug" {
		newInstance = args.SGXInstance
	}
	pi, err := witness.NewPublicInput(guestInput, args.ProofType, args.SGXType, newInstance)
	if err != nil {
		return err
	}
	piHash, err := pi.Hash()
	if err != nil {
		return err
	}

	sign, err := Sign(piHash.Bytes(), prevPrivKey)
	if err != nil {
		return err
	}

	proof := NewOneshotProof(args.SGXInstanceID, newInstance, sign)
	quote, err := provider.LoadQuote(args, newInstance)
	if err != nil {
		return err
	}
	quote.Print()
	return (&ProofResponse{
		Proof:           proof,
		Quote:           quote.Bytes(),
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           piHash,
	}).Output(args.ProofWriter)
}
