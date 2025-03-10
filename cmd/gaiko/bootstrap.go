package main

import (
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

func bootstrap(cli *cli.Context) error {
	args := flags.NewArguments(cli)
	sgxProver := prover.NewSGXProver(args)
	return sgxProver.Bootstrap(cli.Context, args)
}
