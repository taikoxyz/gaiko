package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func aggregate(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSgxProver(args)
	proof, err := sgxProver.Aggregate(context.Background())
	if err != nil {
		return err
	}
	return proof.Stdout()
}
