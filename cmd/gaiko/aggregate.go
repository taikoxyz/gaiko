package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
)

func aggregate(ctx context.Context, sgxProver prover.Prover, args *flags.Arguments) error {
	return sgxProver.Aggregate(ctx, args)
}
