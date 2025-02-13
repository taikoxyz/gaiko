package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/gaiko/internal/transition"
	"github.com/taikoxyz/gaiko/internal/version"
	"github.com/urfave/cli/v2"
)

var oneshotCommand = &cli.Command{
	Name:   "oneshot",
	Usage:  "Run a state transition",
	Action: transition.Oneshot,
	Flags: []cli.Flag{
		transition.OneShotSgxInstanceID,
	},
}

// NewApp creates an app with sane defaults.
func NewApp(usage string) *cli.App {
	git, _ := version.VCS()
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = params.VersionWithCommit(git.Commit, git.Date)
	app.Usage = usage
	app.Copyright = "Copyright 2013-2024 The gaiko Authors"
	return app
}

var app = NewApp("The Gaiko command line interface")

func init() {
	app.Flags = transition.GlobalFlags
	app.Commands = []*cli.Command{
		oneshotCommand,
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
