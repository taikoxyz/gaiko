package flags

import (
	"os"
	"path"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

const defaultGaikoUserConfigSubDir = ".config/gaiko"

var (
	GlobalSecretDir = &cli.StringFlag{
		Name:  "secret-dir",
		Usage: "Directory for the secret files",
	}

	GlobalConfigDir = &cli.StringFlag{
		Name:  "config-dir",
		Usage: "Directory for configuration files",
	}

	GlobalSgxType = &cli.StringFlag{
		Name:    "sgx-type",
		Usage:   `Which SGX type? "debug", "ego" or gramine`,
		EnvVars: []string{"SGX_TYPE"},
	}

	OneShotSgxInstanceID = &cli.Uint64Flag{
		Name:  "sgx-instance-id",
		Usage: "SGX Instance ID for one-shot operation",
	}
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Crit("Get home dir failed", err)
	}
	GlobalSecretDir.DefaultText = path.Join(home, defaultGaikoUserConfigSubDir, "secrets")
	GlobalConfigDir.DefaultText = path.Join(home, defaultGaikoUserConfigSubDir, "config")
}

const (
	DebugSGXType   = "debug"
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
	SGXType    string
	InstanceID uint32
}

func NewArguments(cli *cli.Context) *Arguments {
	return &Arguments{
		SecretDir:  cli.String(GlobalSecretDir.Name),
		ConfigDir:  cli.String(GlobalConfigDir.Name),
		SGXType:    cli.String(GlobalSgxType.Name),
		InstanceID: uint32(cli.Uint64(OneShotSgxInstanceID.Name)),
	}
}
