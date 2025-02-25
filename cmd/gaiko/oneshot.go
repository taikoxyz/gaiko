package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func oneshot(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSgxProver(cli.String(flags.GlobalSgxType.Name), args.SecretDir)
	proof, err := sgxProver.Oneshot(context.Background(), args)

	if err != nil {
		return err
	}
	return proof.Stdout()
}
