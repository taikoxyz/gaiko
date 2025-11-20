package main

import (
	"context"
	"fmt"
	"os"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/taikoxyz/gaiko/internal/version"
	"github.com/urfave/cli/v2"
)

type GaikoActionFunc func(ctx context.Context, sgxProver prover.Prover, args *flags.Arguments) error

func withSGX(
	action GaikoActionFunc,
) cli.ActionFunc {
	return func(cli *cli.Context) error {
		args := flags.NewArguments(cli)
		sgxProver := prover.NewSGXProver(args)
		return action(cli.Context, sgxProver, args)
	}
}

var oneshotCommand = &cli.Command{
	Name:   "one-shot",
	Usage:  "Run state transition once",
	Action: withSGX(oneshot),
	Flags:  flags.ProofFlags,
}

var batchOneshotCommand = &cli.Command{
	Name:   "one-batch-shot",
	Usage:  "Run multi states transition once",
	Action: withSGX(batchOneshot),
	Flags:  flags.ProofFlags,
}

var bootstrapCommand = &cli.Command{
	Name:   "bootstrap",
	Usage:  "Run the bootstrap process",
	Action: withSGX(bootstrap),
	Flags:  flags.BootstrapFlags,
}

var aggregateCommand = &cli.Command{
	Name:   "aggregate",
	Usage:  "Run the aggregate process",
	Action: withSGX(aggregate),
	Flags:  flags.ProofFlags,
}

var shastaAggregateCommand = &cli.Command{
	Name:   "shasta-aggregate",
	Usage:  "Run Shasta aggregate proof",
	Action: withSGX(shastaAggregate),
	Flags:  flags.ProofFlags,
}

var checkCommand = &cli.Command{
	Name:   "check",
	Usage:  "Run the check process",
	Action: withSGX(check),
}

var serverCommand = &cli.Command{
	Name:    "server",
	Aliases: []string{"serve", "s"},
	Usage:   "Start Gaiko HTTP Server",
	Flags:   flags.ServerFlags,
	Action:  runServer,
}

// newApp creates an app with sane defaults.
func newApp(usage string) *cli.App {
	git, _ := version.VCS()
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = version.WithCommit(git.Commit, git.Date)
	app.Usage = usage
	app.Copyright = "Copyright 2025-2025 The Gaiko Authors"
	return app
}

var app = newApp("The Gaiko command line interface")

func init() {
	app.Flags = flags.GlobalFlags
	app.Commands = []*cli.Command{
		oneshotCommand,
		batchOneshotCommand,
		aggregateCommand,
		shastaAggregateCommand,
		bootstrapCommand,
		checkCommand,
		serverCommand,
	}
	app.Before = flags.InitLogger
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
