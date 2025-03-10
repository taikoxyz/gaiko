package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
)

func oneshot(ctx context.Context, sgxProver prover.Prover, args *flags.Arguments) error {
	return sgxProver.Oneshot(ctx, args)
}

func batchOneshot(ctx context.Context, sgxProver prover.Prover, args *flags.Arguments) error {
	return sgxProver.BatchOneshot(ctx, args)
}
