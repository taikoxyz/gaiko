package main

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func oneshot(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSGXProver(args)
	proof, err := sgxProver.Oneshot(context.Background())

	if err != nil {
		return err
	}
	return proof.Stdout()
}

func batchOneshot(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSGXProver(args)
	proof, err := sgxProver.BatchOneshot(context.Background())

	if err != nil {
		return err
	}
	return proof.Stdout()
}
