package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
)

func bootstrap(ctx context.Context, p prover.Prover, args *flags.Arguments) error {
	return p.Bootstrap(ctx, args)
}
