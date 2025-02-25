package prover

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/sgx"
)

type GramineProver struct {
	provider sgx.Provider
}

var _ Prover = (*GramineProver)(nil)

func NewGramineProver(secretDir string) *GramineProver {
	return &GramineProver{
		provider: sgx.NewGramineProvider(secretDir),
	}
}

func (p *GramineProver) Oneshot(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (p *GramineProver) BatchOneshot(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (p *GramineProver) Aggregate(
	ctx context.Context,
	args *flags.Arguments,
) (*ProofResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (p *GramineProver) Bootstrap(ctx context.Context, args *flags.Arguments) error {
	panic("not implemented") // TODO: Implement
}

func (p *GramineProver) Check(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}
