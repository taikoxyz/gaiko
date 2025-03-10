package main

import (
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func check(cli *cli.Context) error {
	sgxProver := prover.NewSGXProver()
	args := flags.NewArguments(cli)
	return sgxProver.Check(cli.Context, args)
}
