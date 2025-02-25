package prover

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/sgx"
	"github.com/taikoxyz/gaiko/internal/witness"
)

type SgxProver struct {
	provider sgx.Provider
	args     *flags.Arguments
}

var _ Prover = (*SgxProver)(nil)

func NewSgxProver(args *flags.Arguments) *SgxProver {
	return &SgxProver{
		args:     args,
		provider: sgx.NewProvider(args),
	}
}

func (p *SgxProver) Oneshot(
	ctx context.Context,
) (*ProofResponse, error) {
	var driver witness.GuestInput
	return genSgxProof(ctx, p.args, &driver, p.provider)
}

func (p *SgxProver) BatchOneshot(
	ctx context.Context,
) (*ProofResponse, error) {
	var driver witness.BatchGuestInput
	return genSgxProof(ctx, p.args, &driver, p.provider)
}

func (p *SgxProver) Aggregate(
	ctx context.Context,
) (*ProofResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (p *SgxProver) Bootstrap(ctx context.Context) error {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	err = p.provider.SavePrivateKey(privKey)
	if err != nil {
		return err
	}
	fmt.Printf("Public key: %x\n", privKey.PublicKey)
	newInstance := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("Instance address: %x\n", newInstance)

	quote, err := p.provider.LoadQuote(newInstance)
	if err != nil {
		return err
	}
	b := &sgx.BootstrapData{
		PublicKey:   crypto.FromECDSAPub(&privKey.PublicKey),
		NewInstance: newInstance,
		Quote:       quote,
	}

	return p.provider.SaveBootstrap(b)
}

func (p *SgxProver) Check(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}
