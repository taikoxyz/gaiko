package main

import (
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func oneshot(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSGXProver(args)
	return sgxProver.Oneshot(cli.Context, args)
}

func batchOneshot(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSGXProver(args)
	return sgxProver.BatchOneshot(cli.Context, args)
}
