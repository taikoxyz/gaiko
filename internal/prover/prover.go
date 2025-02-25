package prover

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/sgx"
	"github.com/taikoxyz/gaiko/internal/transition"
	"github.com/taikoxyz/gaiko/internal/witness"
)

type Prover interface {
	Oneshot(ctx context.Context, args *flags.Arguments) (*ProofResponse, error)
	BatchOneshot(ctx context.Context, args *flags.Arguments) (*ProofResponse, error)
	Aggregate(ctx context.Context, args *flags.Arguments) (*ProofResponse, error)
	Bootstrap(ctx context.Context, args *flags.Arguments) error
	Check(ctx context.Context) error
}

func genProof(
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

	proof := NewProof(args.InstanceID, newInstance, sign)
	if err = provider.SavePublicKey(newInstance); err != nil {
		return nil, err
	}

	quote, err := provider.LoadQuote()
	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		Proof:           hexutil.Bytes(proof[:]),
		Quote:           hexutil.Bytes(quote),
		PublicKey:       hexutil.Bytes(crypto.FromECDSAPub(&prevPrivKey.PublicKey)),
		InstanceAddress: newInstance,
		Input:           piHash,
	}, nil
}
