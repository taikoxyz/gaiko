// Package flags provides command-line flags and argument parsing for the Gaiko client software.
package flags

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/urfave/cli/v2"
)

const (
	defaultUserConfigDir = ".config/raiko"
	globalCategory       = "GLOBAL"
	serverCategory       = "SERVER"
	proofCategory        = "PROOF"
	bootstrapCategory    = "BOOTSTRAP"
	loggingCategory      = "LOGGING"
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

	SGXInstanceIDsFlag = &cli.StringFlag{
		Name:     "sgx-instance-ids",
		Usage:    "SGX Instance IDs mapping in JSON format, e.g. '{\"PACAYA\":1,\"SHASTA\":2}'",
		Category: serverCategory,
		Value:    "{}",
	}

	HTTPListenPortFlag = &cli.StringFlag{
		Name:     "port",
		Usage:    "Port for the server to listen on",
		Category: serverCategory,
		Value:    "8080",
	}

	SGXInstanceIDFlag = &cli.Uint64Flag{
		Name:     "sgx-instance-id",
		Usage:    "SGX Instance ID for one-(batch-)shot operation",
		Category: proofCategory,
	}

	WitnessFlag = &cli.StringFlag{
		Name:     "witness",
		Usage:    "`stdin` or file name of where to find the witness data to use.",
		Value:    stdinSelector,
		Category: proofCategory,
	}

	ProofFlag = &cli.StringFlag{
		Name:     "proof",
		Usage:    "`stdout` or file name of where to write the proof data.",
		Value:    stdoutSelector,
		Category: proofCategory,
	}

	BootstrapFlag = &cli.StringFlag{
		Name:     "bootstrap",
		Usage:    "`stdout` or file name of where to write the bootstrap data.",
		Value:    stdoutSelector,
		Category: bootstrapCategory,
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
	GlobalSecretDirFlag.Value = filepath.Join(home, defaultUserConfigDir, "secrets")
	GlobalConfigDirFlag.Value = filepath.Join(home, defaultUserConfigDir, "config")
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

var ServerFlags = []cli.Flag{
	SGXInstanceIDsFlag,
	HTTPListenPortFlag,
}

var ProofFlags = []cli.Flag{
	SGXInstanceIDFlag,
	WitnessFlag,
	ProofFlag,
}

var BootstrapFlags = []cli.Flag{
	BootstrapFlag,
}

type Arguments struct {
	SecretDir string
	ConfigDir string
	SGXType   string
	ProofType witness.ProofType
	// if SGXType is "debug", specify the SGX instance address with custom private key
	SGXInstance     common.Address
	SGXInstanceID   uint32
	SGXInstanceIDs  map[string]uint32
	WitnessReader   io.Reader
	ProofWriter     io.Writer
	BootstrapWriter io.Writer
}

func (args *Arguments) Copy() *Arguments {
	argsCopy := *args
	return &argsCopy
}

// UpdateSGXInstanceID updates the SGXInstanceID field based on the provided hardFork name
func (args *Arguments) UpdateSGXInstanceID(hardFork string) {
	if args.SGXInstanceID != 0 {
		return
	}
	hardFork = strings.ToUpper(hardFork)
	args.SGXInstanceID = args.SGXInstanceIDs[hardFork]
}

func NewArguments(cli *cli.Context) *Arguments {
	var (
		err             error
		secretDir       = cli.String(GlobalSecretDirFlag.Name)
		configDir       = cli.String(GlobalConfigDirFlag.Name)
		witnessReader   io.Reader
		witnessStr      = cli.String(WitnessFlag.Name)
		proofWriter     io.Writer
		proofStr        = cli.String(ProofFlag.Name)
		bootstrapStr    = cli.String(BootstrapFlag.Name)
		bootstrapWriter io.Writer
	)
	if witnessStr == stdinSelector || witnessStr == "" {
		witnessReader = os.Stdin
	} else {
		if witnessReader, err = os.Open(witnessStr); err != nil {
			panic(err)
		}
	}

	if proofStr == stdoutSelector || proofStr == "" {
		proofWriter = os.Stdout
	} else {
		if proofWriter, err = os.Create(proofStr); err != nil {
			panic(err)
		}
	}
	if bootstrapStr == stdoutSelector || bootstrapStr == "" {
		bootstrapWriter = os.Stdout
	} else {
		if bootstrapWriter, err = os.Create(bootstrapStr); err != nil {
			panic(err)
		}
	}

	// Parse SGX Instance IDs from JSON
	sgxInstanceIDsStr := cli.String(SGXInstanceIDsFlag.Name)
	sgxInstanceIDs := make(map[string]uint32)
	if sgxInstanceIDsStr != "" && sgxInstanceIDsStr != "{}" {
		if err := json.Unmarshal([]byte(sgxInstanceIDsStr), &sgxInstanceIDs); err != nil {
			panic(err)
		}
	}

	return &Arguments{
		SecretDir:       secretDir,
		ConfigDir:       configDir,
		SGXType:         cli.String(GlobalSGXTypeFlag.Name),
		ProofType:       witness.SGXGethProofType,
		SGXInstanceID:   uint32(cli.Uint64(SGXInstanceIDFlag.Name)),
		SGXInstanceIDs:  sgxInstanceIDs,
		WitnessReader:   witnessReader,
		ProofWriter:     proofWriter,
		BootstrapWriter: bootstrapWriter,
	}
}

// InitLogger initializes the root logger with the command line flags.
func InitLogger(c *cli.Context) error {
	slogVerbosity := log.FromLegacyLevel(c.Int(VerbosityFlag.Name))
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
