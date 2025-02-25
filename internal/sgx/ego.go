package sgx

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
)

type EgoProver struct{}

func (p *EgoProver) Oneshot(ctx context.Context, args *flags.Arguments) (*Proof, error) {
	panic("not implemented") // TODO: Implement
}

func (p *EgoProver) Aggregate(ctx context.Context, args *flags.Arguments) (*Proof, error) {
	panic("not implemented") // TODO: Implement
}

func (p *EgoProver) Bootstrap(ctx context.Context, args *flags.Arguments) error {
	panic("not implemented") // TODO: Implement
}

func (p *EgoProver) Check(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}
