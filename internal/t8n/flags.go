package t8n

import "github.com/urfave/cli/v2"

var (
	GlobalSecretPath = &cli.StringFlag{
		Name:  "secret.path",
		Usage: "Path to the secret file",
	}

	GlobalConfigDir = &cli.StringFlag{
		Name:  "config.dir",
		Usage: "Directory for configuration files",
	}

	OneShotSgxInstanceID = &cli.Uint64Flag{
		Name:  "oneshot.sgxInstanceID",
		Usage: "SGX Instance ID for one-shot operation",
	}
)

var GlobalFlags = []cli.Flag{
	GlobalSecretPath,
	GlobalConfigDir,
}
