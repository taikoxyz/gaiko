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

const (
	stdinSelector  = "stdin"
	stdoutSelector = "stdout"
)

var (
	GlobalSecretDirFlag = &cli.StringFlag{
		Name:     "secret-dir",
		Usage:    "Directory for the secret files",
		Category: globalCategory,
	}

	GlobalConfigDirFlag = &cli.StringFlag{
		Name:     "config-dir",
		Usage:    "Directory for configuration files",
		Category: globalCategory,
	}

	GlobalSGXTypeFlag = &cli.StringFlag{
		Name:     "sgx-type",
		Usage:    `Which SGX type? "debug", "ego" or "gramine"`,
		Category: globalCategory,
		EnvVars:  []string{"SGX_TYPE"},
	}

	SGXInstanceIDFlag = &cli.Uint64Flag{
		Name:  "sgx-instance-id",
		Usage: "SGX Instance ID for one-(batch-)shot operation",
	}

	WitnessFlag = &cli.StringFlag{
		Name:  "witness",
		Usage: "`stdin` or file name of where to find the witness data to use.",
		Value: stdinSelector,
	}

	ProofFlag = &cli.StringFlag{
		Name:  "proof",
		Usage: "`stdout` or file name of where to write the proof data.",
		Value: stdoutSelector,
	}

	BootstrapFlag = &cli.StringFlag{
		Name:  "bootstrap",
		Usage: "`stdout` or file name of where to write the bootstrap data.",
		Value: stdoutSelector,
	}

	// Optional flags used by all client software.
	// Logging
	VerbosityFlag = &cli.IntFlag{
		Name:     "verbosity",
		Usage:    "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value:    3,
		Category: loggingCategory,
		EnvVars:  []string{"VERBOSITY"},
	}
	LogJSONFlag = &cli.BoolFlag{
		Name:     "log-json",
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
	GlobalSecretDirFlag.DefaultText = filepath.Join(home, defaultGaikoUserConfigSubDir, "secrets")
	GlobalConfigDirFlag.DefaultText = filepath.Join(home, defaultGaikoUserConfigSubDir, "config")
}

const (
	EgoSGXType     = "ego"
	GramineSGXType = "gramine"
)

var GlobalFlags = []cli.Flag{
	GlobalSecretDirFlag,
	GlobalConfigDirFlag,
	GlobalSGXTypeFlag,
	VerbosityFlag,
	LogJSONFlag,
}

type Arguments struct {
	SecretDir string
	ConfigDir string
	SGXType   string
	ProofType witness.ProofType
	// if SGXType is "debug", specify the SGX instance address with custom private key
	SGXInstance     common.Address
	SGXInstanceID   uint32
	WitnessReader   io.Reader
	ProofWriter     io.Writer
	BootstrapWriter io.Writer
}

func (args *Arguments) Copy() *Arguments {
	argsCopy := *args
	return &argsCopy
}

func NewArguments(cli *cli.Context) *Arguments {
	secretDir := GlobalSecretDirFlag.GetDefaultText()
	if cli.IsSet(GlobalSecretDirFlag.Name) {
		secretDir = cli.String(GlobalSecretDirFlag.Name)
	}
	configDir := GlobalConfigDirFlag.GetDefaultText()
	if cli.IsSet(GlobalConfigDirFlag.Name) {
		configDir = cli.String(GlobalConfigDirFlag.Name)
	}
	var (
		err             error
		witnessReader   io.Reader
		witnessStr      = cli.String(WitnessFlag.Name)
		proofWriter     io.Writer
		proofStr        = cli.String(ProofFlag.Name)
		bootstrapStr    = cli.String(BootstrapFlag.Name)
		bootstrapWriter io.Writer
	)
	if witnessStr != stdinSelector {
		witnessReader = os.Stdin
	} else {
		if witnessReader, err = os.Open(witnessStr); err != nil {
			panic(err)
		}
	}

	if proofStr != stdoutSelector {
		proofWriter = os.Stdout
	} else {
		if proofWriter, err = os.Create(proofStr); err != nil {
			panic(err)
		}
	}
	if bootstrapStr != stdoutSelector {
		bootstrapWriter = os.Stdout
	} else {
		if bootstrapWriter, err = os.Create(bootstrapStr); err != nil {
			panic(err)
		}
	}
	return &Arguments{
		SecretDir:       secretDir,
		ConfigDir:       configDir,
		SGXType:         cli.String(GlobalSGXTypeFlag.Name),
		ProofType:       witness.SGXGethProofType,
		SGXInstanceID:   uint32(cli.Uint64(SGXInstanceIDFlag.Name)),
		WitnessReader:   witnessReader,
		ProofWriter:     proofWriter,
		BootstrapWriter: bootstrapWriter,
	}
}

// InitLogger initializes the root logger with the command line flags.
func InitLogger(c *cli.Context) error {
	var (
		slogVerbosity = log.FromLegacyLevel(c.Int(VerbosityFlag.Name))
	)
	if c.Bool(LogJSONFlag.Name) {
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
