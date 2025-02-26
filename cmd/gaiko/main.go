package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/version"
	"github.com/urfave/cli/v2"
)

var oneshotCommand = &cli.Command{
	Name:   "one-shot",
	Usage:  "Run state transition once",
	Action: oneshot,
	Flags: []cli.Flag{
		flags.SGXInstanceID,
	},
}

var batchOneshotCommand = &cli.Command{
	Name:   "one-batch-shot",
	Usage:  "Run multi states transition once",
	Action: batchOneshot,
	Flags: []cli.Flag{
		flags.SGXInstanceID,
	},
}

var bootstrapCommand = &cli.Command{
	Name:   "bootstrap",
	Usage:  "Run the bootstrap process",
	Action: bootstrap,
}

var aggregateCommand = &cli.Command{
	Name:   "aggregate",
	Usage:  "Run the aggregate process",
	Action: aggregate,
}

var checkCommand = &cli.Command{
	Name:   "check",
	Usage:  "Run the check process",
	Action: check,
}

// newApp creates an app with sane defaults.
func newApp(usage string) *cli.App {
	git, _ := version.VCS()
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = params.VersionWithCommit(git.Commit, git.Date)
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
		bootstrapCommand,
		checkCommand,
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
