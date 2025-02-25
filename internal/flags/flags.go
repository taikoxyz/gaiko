package flags

import "github.com/urfave/cli/v2"

var (
	GlobalSecretDir = &cli.StringFlag{
		Name:  "global.secretDir",
		Usage: "Directory for the secret files",
	}

	GlobalConfigDir = &cli.StringFlag{
		Name:  "global.configDir",
		Usage: "Directory for configuration files",
	}

	GlobalSgxType = &cli.StringFlag{
		Name:  "global.sgxType",
		Usage: "Which SGX type? ego or gramine",
	}

	OneShotSgxInstanceID = &cli.Uint64Flag{
		Name:  "oneshot.sgxInstanceID",
		Usage: "SGX Instance ID for one-shot operation",
	}
)

const (
	EgoSGXType     = "ego"
	GramineSGXType = "gramine"
)

var GlobalFlags = []cli.Flag{
	GlobalSecretDir,
	GlobalConfigDir,
	GlobalSgxType,
}

type Arguments struct {
	SecretDir  string
	ConfigDir  string
	SgxType    string
	InstanceID uint32
}

func NewArguments(cli *cli.Context) *Arguments {
	return &Arguments{
		SecretDir:  cli.String(GlobalSecretDir.Name),
		ConfigDir:  cli.String(GlobalConfigDir.Name),
		SgxType:    cli.String(GlobalSgxType.Name),
		InstanceID: uint32(cli.Uint64(OneShotSgxInstanceID.Name)),
	}
}
