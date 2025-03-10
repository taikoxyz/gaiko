package prover

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
)

type Prover interface {
	Oneshot(ctx context.Context, args *flags.Arguments) error
	BatchOneshot(ctx context.Context, args *flags.Arguments) error
	Aggregate(ctx context.Context, args *flags.Arguments) error
	Bootstrap(ctx context.Context, args *flags.Arguments) error
	Check(ctx context.Context, args *flags.Arguments) error
}
