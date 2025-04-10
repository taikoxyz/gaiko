package flags

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/urfave/cli/v2"
)

const (
	defaultGaikoUserConfigSubDir = ".config/raiko"
	globalCategory               = "GLOBAL"
	loggingCategory              = "LOGGING"
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

	// Optional flags used by all client software.
	// Logging
	Verbosity = &cli.IntFlag{
		Name:     "verbosity",
		Usage:    "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value:    3,
		Category: loggingCategory,
		EnvVars:  []string{"VERBOSITY"},
	}
	LogJSON = &cli.BoolFlag{
		Name:     "log.json",
		Usage:    "Format logs with JSON",
		Category: loggingCategory,
		EnvVars:  []string{"LOG_JSON"},
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
	Verbosity,
	LogJSON,
}

type Arguments struct {
	SecretDir       string
	ConfigDir       string
	SGXType         string
	ProverType      witness.ProofType
	SgxInstance     common.Address
	SGXInstanceID   uint32
	WitnessReader   io.Reader
	ProofWriter     io.Writer
	BootstrapWriter io.Writer
}

func NewArguments(cli *cli.Context) *Arguments {
	secretDir := GlobalSecretDir.GetDefaultText()
	if cli.IsSet(GlobalSecretDir.Name) {
		secretDir = cli.String(GlobalSecretDir.Name)
	}
	configDir := GlobalConfigDir.GetDefaultText()
	if cli.IsSet(GlobalConfigDir.Name) {
		configDir = cli.String(GlobalConfigDir.Name)
	}
	return &Arguments{
		SecretDir:       secretDir,
		ConfigDir:       configDir,
		SGXType:         cli.String(GlobalSGXType.Name),
		ProverType:      witness.SGXGethProofType,
		SGXInstanceID:   uint32(cli.Uint64(SGXInstanceID.Name)),
		WitnessReader:   os.Stdin,
		ProofWriter:     os.Stdout,
		BootstrapWriter: os.Stdout,
	}
}

// InitLogger initializes the root logger with the command line flags.
func InitLogger(c *cli.Context) error {
	var (
		slogVerbosity = log.FromLegacyLevel(c.Int(Verbosity.Name))
	)
	if c.Bool(LogJSON.Name) {
		glogger := log.NewGlogHandler(log.NewGlogHandler(log.JSONHandler(os.Stdout)))
		glogger.Verbosity(slogVerbosity)
		log.SetDefault(log.NewLogger(glogger))
	} else {
		glogger := log.NewGlogHandler(log.NewTerminalHandler(os.Stdout, false))
		glogger.Verbosity(slogVerbosity)
		log.SetDefault(log.NewLogger(glogger))
	}
	return nil
}
