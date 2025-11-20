# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gaiko is a Golang port of Raiko, a state transition prover for Taiko blockchain that runs in Trusted Execution Environments (TEEs). It performs state transitions for Ethereum-compatible blocks and generates cryptographic proofs using SGX (Software Guard Extensions) technology.

## Quick Start

### Development Setup

```bash
# Build for development (no TEE required)
go build -tags dev -o gaiko ./cmd/gaiko

# Run tests with required environment variable
GAIKO=1 go test ./...

# Run specific test suite
cd tests && GAIKO=1 go test -v -run TestSingle
```

### Common Workflows

**Running a single block proof:**

```bash
GAIKO=1 ./gaiko one-shot --sgx-type dev --witness input.json --proof output.json
```

**Running batch block proofs:**

```bash
GAIKO=1 ./gaiko one-batch-shot --sgx-type dev --witness batch-input.json --proof batch-output.json
```

**Starting the API server:**

```bash
GAIKO=1 ./gaiko server --sgx-type dev --port 8080
```

## Build Commands

### Standard Build

```bash
go build -o gaiko ./cmd/gaiko
```

### SGX/TEE Builds

For production SGX builds using EGO:

```bash
ego-go build -o gaiko-ego ./cmd/gaiko
ego sign && ego bundle gaiko-ego gaiko
ego uniqueid gaiko-ego
ego signerid gaiko-ego
```

For development mode (without TEE):

```bash
go build -tags dev -o gaiko ./cmd/gaiko
```

### Docker Build

```bash
./scripts/build-docker.sh
# Interactive: choose local(0) for gaiko-local image
```

## Testing

**IMPORTANT**: All tests require the `GAIKO=1` environment variable to be set.

### Run All Tests

```bash
GAIKO=1 go test ./...
```

### Run Specific Test Suites

```bash
cd tests
GAIKO=1 go test -v -run TestSingle    # Test single block transitions
GAIKO=1 go test -v -run TestBatch     # Test batch block transitions
GAIKO=1 go test -v -run TestAggregate # Test aggregate proofs
```

### Run Tests in Specific Packages

```bash
GAIKO=1 go test -v ./internal/prover/...
GAIKO=1 go test -v ./internal/transition/...
GAIKO=1 go test -v ./pkg/mpt/...
```

### Run Tests with Race Detection

```bash
GAIKO=1 go test -race ./...
```

### Run Tests with Coverage

```bash
GAIKO=1 go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Environment Variables

- `GAIKO=1` - **Required** for all tests and runtime execution. Must be set when running gaiko commands.
- `PORT` - Override default server port (8080)
- `LOG_LEVEL` - Control logging verbosity (debug, info, warn, error)

## CLI Commands

Gaiko provides several subcommands:

- `one-shot` - Run state transition for a single block
- `one-batch-shot` - Run state transitions for multiple blocks in a batch
- `aggregate` - Run aggregate proof generation
- `bootstrap` - Initialize TEE instance with cryptographic keys
- `check` - Verify TEE instance configuration
- `server` - Start HTTP API server (port 8080 by default)

### Common Flags

All commands support:

- `--sgx-type` - TEE type: "debug", "ego", "gramine", or "dev"
- `--witness` - Input witness file path
- `--proof` - Output proof file path
- `--config-dir` - Configuration directory (default: ~/.config/raiko/config)
- `--secret-dir` - Secrets directory (default: ~/.config/raiko/secrets)

### CLI Examples

**Initialize TEE instance:**

```bash
GAIKO=1 ./gaiko bootstrap --sgx-type ego --config-dir ./config --secret-dir ./secrets
```

**Check TEE configuration:**

```bash
GAIKO=1 ./gaiko check --sgx-type ego
```

**Generate aggregate proof:**

```bash
GAIKO=1 ./gaiko aggregate --sgx-type dev --witness aggregate-input.json --proof aggregate-output.json
```

## Architecture

### High-Level Data Flow

1. **Input**: Witness data (GuestInput or BatchGuestInput) containing block headers, transactions, state tries, and Taiko-specific metadata
2. **Verification**: Validate chain specs, blob proofs, and input consistency
3. **Transition**: Execute block transactions using stateless witness approach
4. **Proof Generation**: Generate SGX attestation quotes and cryptographic proofs
5. **Output**: ProofResponse with state root, block hash, and TEE attestation

### Core Components

**`internal/prover`** - Prover abstraction layer

- `Prover` interface defines methods: Oneshot, BatchOneshot, Aggregate, Bootstrap, Check
- `SGXProver` implements TEE-backed proving using Provider interface
- Proof generation (genOneshotProof, genAggregateProof) handles signing and attestation

**`internal/witness`** - Witness data structures

- `GuestInput`: Single block witness with parent state, transactions, Taiko metadata
- `BatchGuestInput`: Multiple block witness for batch processing
- `WitnessInput` interface with methods: GuestInputs(), Verify(), BlockMetadataFork(), etc.
- Implements unmarshaling logic for different Taiko hard forks (Hekla, Ontake, Shasta)
- Converts MPT tries to stateless witnesses for go-ethereum's ExecuteStateless

**`internal/transition`** - State transition execution

- `ExecuteAndVerify()`: Main entry point that executes blocks using stateless witness
- Uses `core.ExecuteStateless()` from go-ethereum for verification
- Validates state root and receipt root match expected values
- Runs transitions concurrently using errgroup for batch inputs

**`internal/tee`** - TEE provider abstraction

- `Provider` interface for SGX operations (LoadQuote, LoadPrivateKey, SavePrivateKey)
- Implementations: EGO (`sgx_ego.go`), Gramine (`sgx_gramine.go`), TDX (`tdx.go`), Dev mode (`dev.go`)
- Bootstrap data saves public key, instance address, and attestation quote

**`internal/types`** - Type conversions between Raiko and Gaiko

- Ethereum types (Header, Transaction, Block, AccessList)
- Taiko event types (BlockProposed, BlockMetadata, BatchProposed)
- Uses `fjl/gencodec` for JSON marshaling code generation

**`pkg/mpt`** - Merkle Patricia Trie implementation

- Ported from Raiko's Rust implementation
- Provides stateful MPT operations (Get, Insert, Delete, Hash)
- Converts to stateless witness format for go-ethereum

**`pkg/keccak`** - Pooled Keccak hasher

- Thread-safe hasher pool for performance
- Single function `Keccak(data []byte) common.Hash`

### Hard Fork Support

Gaiko supports multiple Taiko protocol versions:

- **Hekla**: Uses BlockMetadata with single block parameters
- **Ontake**: Uses BlockMetadataV2 with enhanced fields (liveness bond, base fee config)
- **Pacaya**: Batch mode with multi-block support
- **Shasta**: Latest fork with manifest-based transaction derivation

Hard fork detection is automatic based on BlockProposed event structure.

### Witness to Stateless Conversion

The key innovation is `GuestInput.NewWitness()` in `internal/witness/stateless.go`:

1. Collects all MPT nodes from parent state trie and storage tries
2. Validates parent state root matches expected value
3. Packages nodes into `stateless.Witness` structure
4. Allows go-ethereum's `core.ExecuteStateless()` to verify transitions without full state DB

### Build Tags

- `dev` tag: Use development TEE provider (no actual SGX, for testing)
- Default (no tags): Use production TEE providers (EGO/Gramine/TDX)

### Testing Fixtures

Test fixtures in `tests/fixtures/` contain:

- JSON input files with witness data
- JSON output files with expected proofs
- Embedded using `//go:embed` for convenient testing

## Code Generation

Several types use code generation:

- `internal/types/gen_*.go` - JSON marshalers via `fjl/gencodec`
- `internal/witness/gen_*.go` - Chain spec and constant marshalers

Regenerate with:

```bash
go generate ./...
```

## Dependencies

Key dependencies:

- `github.com/ethereum/go-ethereum` - Core Ethereum types and execution (custom fork: taikoxyz/taiko-geth)
- `github.com/ethereum-optimism/optimism` - Blob utilities (custom fork: taikoxyz/optimism)
- `github.com/taikoxyz/taiko-mono` - Taiko protocol bindings and utilities
- `github.com/edgelesssys/ego` - EGO SGX framework
- `github.com/google/go-tdx-guest` - TDX attestation
- `github.com/urfave/cli/v2` - CLI framework

## Debugging

### Enable Debug Logging

```bash
GAIKO=1 LOG_LEVEL=debug ./gaiko one-shot --sgx-type dev --witness input.json
```

### Run with Delve Debugger

```bash
GAIKO=1 dlv debug ./cmd/gaiko -- one-shot --sgx-type dev --witness input.json
```

### Useful Debug Commands

```bash
# Check Go environment
go env

# Verify module dependencies
go mod verify
go mod tidy

# View test output with verbose logging
GAIKO=1 go test -v -count=1 ./tests
```

## Troubleshooting

### Common Issues

**Commands fail or behave unexpectedly:**

- Solution: Always set `GAIKO=1` environment variable for ALL gaiko operations (tests and runtime)
- Example for tests: `GAIKO=1 go test ./...`
- Example for runtime: `GAIKO=1 ./gaiko one-shot --sgx-type dev --witness input.json`

**Build fails with SGX errors:**

- Ensure you're using the correct build command for your target
- For development without TEE: `go build -tags dev -o gaiko ./cmd/gaiko`
- For production SGX: Use `ego-go build` commands

**Witness validation errors:**

- Verify the witness file matches the expected hard fork (Hekla, Ontake, Pacaya, Shasta)
- Check that parent state root is correct
- Ensure block continuity for batch inputs

**State root mismatch:**

- Verify all MPT nodes are included in the witness
- Check transaction execution order matches expected sequence
- Validate chain spec parameters match the network

## Development Tips

### Working with Different Hard Forks

The codebase automatically detects hard forks based on witness structure. When adding support for new forks:

1. Update `internal/witness/input.go` with new metadata structures
2. Add unmarshaling logic in `UnmarshalJSON` methods
3. Update `BlockMetadataFork()` to detect the new fork
4. Add test fixtures in `tests/fixtures/`

### Adding New Test Cases

1. Generate witness data from Raiko or create manually
2. Place in `tests/fixtures/proposals/` directory
3. Add expected output in `tests/fixtures/proofs/`
4. Update test files to include new case
5. Run with: `cd tests && GAIKO=1 go test -v -run TestSingle`

### Code Style Guidelines

- Follow standard Go conventions (gofmt, golint)
- Use meaningful variable names that reflect domain concepts
- Add comments for complex cryptographic or TEE operations
- Keep functions focused and modular
- Write tests for all new functionality

## Important Notes

- **`GAIKO=1` environment variable is required for ALL operations** - both tests and runtime execution. Always prefix commands with `GAIKO=1`
- Always use `go-ethereum`'s stateless witness approach (`core.ExecuteStateless`) for block verification
- TEE providers must implement secure key storage - private keys are sealed to MRENCLAVE
- Blob proofs support two modes: KZG versioned hash or proof of equivalence
- Batch inputs require strict block continuity validation (sequential parent-child relationships)
- For Shasta fork, transaction manifests are decoded from blob data at specific offsets
- This requirement is critical for CI/CD pipelines and production deployments
