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

	GlobalSGXType = &cli.StringFlag{
		Name:    "sgx-type",
		Usage:   `Which SGX type? "debug", "ego" or "gramine"`,
		EnvVars: []string{"SGX_TYPE"},
	}

	SGXInstanceID = &cli.Uint64Flag{
		Name:  "sgx-instance-id",
		Usage: "SGX Instance ID for one-(batch-)shot operation",
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
	GlobalSGXType,
}

type Arguments struct {
	SecretDir     string
	ConfigDir     string
	SGXType       string
	SGXInstanceID uint32
}

func NewArguments(cli *cli.Context) *Arguments {
	return &Arguments{
		SecretDir:     cli.String(GlobalSecretDir.Name),
		ConfigDir:     cli.String(GlobalConfigDir.Name),
		SGXType:       cli.String(GlobalSGXType.Name),
		SGXInstanceID: uint32(cli.Uint64(SGXInstanceID.Name)),
	}
}
