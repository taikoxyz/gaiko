package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
)

func oneshot(ctx context.Context, p prover.Prover, args *flags.Arguments) error {
	return p.Oneshot(ctx, args)
}

func batchOneshot(ctx context.Context, p prover.Prover, args *flags.Arguments) error {
	return p.BatchOneshot(ctx, args)
}
