package flags

import "github.com/urfave/cli/v2"

var (
	GlobalSecretDir = &cli.StringFlag{
		Name:  "secret.dir",
		Usage: "Directory for the secret files",
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
	GlobalSecretDir,
	GlobalConfigDir,
}

type Arguments struct {
	SecretDir  string
	ConfigDir  string
	InstanceID uint32
}

func NewArguments(ctx *cli.Context) *Arguments {
	return &Arguments{
		SecretDir:  ctx.String(GlobalSecretDir.Name),
		ConfigDir:  ctx.String(GlobalConfigDir.Name),
		InstanceID: uint32(ctx.Uint64(OneShotSgxInstanceID.Name)),
	}
}
