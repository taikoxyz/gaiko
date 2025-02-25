package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func oneshot(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	var sgxProver prover.Prover
	switch cli.String(flags.GlobalSgxType.Name) {
	case flags.GramineSGXType:
		sgxProver = prover.NewGramineProver(args.SecretDir)
	default:
		sgxProver = prover.NewEgoProver(args.SecretDir)
	}
	proof, err := sgxProver.Oneshot(context.Background(), args)

	if err != nil {
		return err
	}
	return proof.Stdout()
}
