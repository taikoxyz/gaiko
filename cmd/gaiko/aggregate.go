package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
)

func aggregate(ctx context.Context, p prover.Prover, args *flags.Arguments) error {
	return p.Aggregate(ctx, args)
}
