package flags

import (
	"os"
	"path/filepath"

	"github.com/taikoxyz/taiko-mono/packages/taiko-client/cmd/flags"
	"github.com/urfave/cli/v2"
)

const (
	defaultGaikoUserConfigSubDir = ".config/gaiko"
	globalCategory               = "GLOBAL"
)

var (
	GlobalSecretDir = &cli.StringFlag{
		Name:     "secret-dir",
		Usage:    "Directory for the secret files",
		Category: globalCategory,
	}

	GlobalConfigDir = &cli.StringFlag{
		Name:     "config-dir",
		Usage:    "Directory for configuration files",
		Category: globalCategory,
	}

	GlobalSGXType = &cli.StringFlag{
		Name:     "sgx-type",
		Usage:    `Which SGX type? "debug", "ego" or "gramine"`,
		Category: globalCategory,
		EnvVars:  []string{"SGX_TYPE"},
	}

	SGXInstanceID = &cli.Uint64Flag{
		Name:  "sgx-instance-id",
		Usage: "SGX Instance ID for one-(batch-)shot operation",
	}
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	GlobalSecretDir.DefaultText = filepath.Join(home, defaultGaikoUserConfigSubDir, "secrets")
	GlobalConfigDir.DefaultText = filepath.Join(home, defaultGaikoUserConfigSubDir, "config")
}

const (
	EgoSGXType     = "ego"
	GramineSGXType = "gramine"
)

var GlobalFlags = []cli.Flag{
	GlobalSecretDir,
	GlobalConfigDir,
	GlobalSGXType,
	flags.Verbosity,
	flags.LogJSON,
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
