package prover

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
)

type Prover interface {
	Oneshot(ctx context.Context, args *flags.Arguments) (*ProofResponse, error)
	BatchOneshot(ctx context.Context, args *flags.Arguments) (*ProofResponse, error)
	Aggregate(ctx context.Context, args *flags.Arguments) (*ProofResponse, error)
	Bootstrap(ctx context.Context, args *flags.Arguments) error
	Check(ctx context.Context) error
}
