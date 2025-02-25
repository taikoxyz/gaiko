package prover

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/sgx"
	"github.com/taikoxyz/gaiko/internal/witness"
)

type SgxProver struct {
	provider sgx.Provider
}

var _ Prover = (*SgxProver)(nil)

func NewSgxProver(typ, secretDir string) *SgxProver {
	return &SgxProver{
		provider: sgx.NewProvider(typ, secretDir),
	}
}

func (p *SgxProver) Oneshot(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	var driver witness.GuestInput
	return genSgxProof(ctx, args, &driver, p.provider)
}

func (p *SgxProver) BatchOneshot(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	var driver witness.BatchGuestInput
	return genSgxProof(ctx, args, &driver, p.provider)
}

func (p *SgxProver) Aggregate(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (p *SgxProver) Bootstrap(ctx context.Context, args *flags.Arguments) error {
	panic("not implemented") // TODO: Implement
	// func BootStrap(ctx *cli.Context) error {
	// 	privKey, err := crypto.GenerateKey()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	privKeyPath := ctx.String(flags.GlobalSecretDir.Name)
	// 	err = util.SavePrivKey(privKeyPath, privKey)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("Public key: %x\n", privKey.PublicKey)
	// 	newInstance := crypto.PubkeyToAddress(privKey.PublicKey)
	// 	fmt.Printf("Instance address: %x\n", newInstance)

	// 	return nil
	// }

}

func (p *SgxProver) Check(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}
