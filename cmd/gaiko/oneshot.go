package main

import (
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func oneshot(cli *cli.Context) error {
	sgxProver := prover.NewSGXProver()
	args := flags.NewArguments(cli)
	return sgxProver.Oneshot(cli.Context, args)
}

func batchOneshot(cli *cli.Context) error {
	sgxProver := prover.NewSGXProver()
	args := flags.NewArguments(cli)
	return sgxProver.BatchOneshot(cli.Context, args)
}
