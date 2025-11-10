package prover

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/tee"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/taikoxyz/gaiko/pkg/keccak"
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
	var input witness.ShastaAggregationGuestInput

	err = json.NewDecoder(args.WitnessReader).Decode(&input)
	if err != nil {
		return err
	}

	log.Info("Received Shasta aggregation input", "proposals", len(input.Proposals))

	if len(input.Proposals) == 0 {
		return fmt.Errorf("no proposals to aggregate")
	}

	// Aggregate proposal hashes
	combinedHashes := make([]byte, 0, len(input.Proposals)*common.HashLength)
	for _, proposal := range input.Proposals {
		// Hash each proposal's checkpoint
		checkpointData := make([]byte, 0, 96)
		checkpointData = append(checkpointData, common.BigToHash(common.Big0.SetUint64(proposal.Checkpoint.BlockNumber)).Bytes()...)
		checkpointData = append(checkpointData, proposal.Checkpoint.BlockHash.Bytes()...)
		checkpointData = append(checkpointData, proposal.Checkpoint.StateRoot.Bytes()...)

		proposalHash := keccak.Keccak(checkpointData)
		combinedHashes = append(combinedHashes, proposalHash.Bytes()...)
	}

	aggHash := keccak.Keccak(combinedHashes)
	sign, err := Sign(aggHash.Bytes(), prevPrivKey)
	if err != nil {
		return err
	}

	// For Shasta aggregate, we use a similar structure to regular aggregate
	// but with different semantics
	oldInstance := newInstance // For Shasta, old and new instance are the same
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
