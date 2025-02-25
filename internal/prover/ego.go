package prover

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/sgx"
	"github.com/taikoxyz/gaiko/internal/witness"
)

type EgoProver struct {
	provider sgx.Provider
}

var _ Prover = (*EgoProver)(nil)

func NewEgoProver(secretDir string) *EgoProver {
	return &EgoProver{
		provider: sgx.NewEgoProvider(secretDir),
	}
}

func (p *EgoProver) Oneshot(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	var driver witness.GuestInput
	return genProof(ctx, args, &driver, p.provider)
}

func (p *EgoProver) BatchOneshot(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	var driver witness.BatchGuestInput
	return genProof(ctx, args, &driver, p.provider)
}

func (p *EgoProver) Aggregate(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (p *EgoProver) Bootstrap(ctx context.Context, args *flags.Arguments) error {
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

func (p *EgoProver) Check(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}
