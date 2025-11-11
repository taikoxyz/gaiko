package prover

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/tee"
	"github.com/taikoxyz/gaiko/internal/witness"
)

// genShastaAggregateProof generates an aggregate proof for multiple Shasta proposals.
func genShastaAggregateProof(
	_ context.Context,
	args *flags.Arguments,
	provider tee.Provider,
) error {
	args.UpdateSGXInstanceID(witness.ShastaHardFork)
	prevPrivKey, err := provider.LoadPrivateKey(args)
	if err != nil {
		return err
	}

	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	var input witness.ShastaRawAggregationGuestInput

	err = json.NewDecoder(args.WitnessReader).Decode(&input)
	if err != nil {
		return err
	}
	log.Info("receive input: ", "input", input)
	oldInstance := common.BytesToAddress(input.Proofs[0].Proof[4:24])
	curInstance := oldInstance
	transitionHashes := make([]common.Hash, 0, len(input.Proofs))
	for i, proof := range input.Proofs {
		pubKey, err := SigToPub(proof.Input.Bytes(), proof.Proof[24:])
		if err != nil {
			return err
		}
		if crypto.PubkeyToAddress(*pubKey) != curInstance {
			return fmt.Errorf("invalid proof[%d]", i)
		}
		curInstance = common.BytesToAddress(proof.Proof[4:24])
		transitionHashes = append(transitionHashes, proof.Input)
	}
	if newInstance != curInstance {
		return fmt.Errorf("invalid instance: %#x", curInstance)
	}

	aggHash := witness.HashShastaAggregation(transitionHashes, input.ChainID, input.VerifierAddress, newInstance)
	sign, err := Sign(aggHash.Bytes(), prevPrivKey)
	if err != nil {
		return err
	}

	proof := NewShastaAggregateProof(args.SGXInstanceID, newInstance, sign)
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

func NewShastaAggregateProof(
	instanceID uint32,
	newInstance common.Address,
	sign []byte,
) []byte {
	var proof [89]byte
	binary.BigEndian.PutUint32(proof[:4], instanceID)
	copy(proof[4:24], newInstance.Bytes())
	copy(proof[24:], sign)
	return proof[:]
}
