# Gaiko is Golang port of Raiko

## Architecture

```
.
├── cmd
│   └── gaiko
├── ego
├── internal
│   ├── flags
│   ├── keccak
│   ├── mpt
│   ├── prover
│   ├── sgx
│   ├── transition
│   ├── types
│   ├── version
│   └── witness
└── test
    └── fixtures
        ├── batch
        └── single
```

- `ego` artifacts, e.g. config
- `internal/flags` cli arguments
- `internal/keccak` keccak hasher
- `internal/mpt` the golang porting of merkle trie from `raiko`
- `internal/prover` prover abstracts, the main entry of `gaiko`
- `internal/sgx` sgx provider, `gramine` or `ego`
- `internal/transition` state transition of blocks
- `internal/types` type bridge between `raiko` and `gaiko`
- `witness` proof witness, aka: `Batch/GuestInput` in `raiko`

## Command Help

```
$ ./gaiko
NAME:
   gaiko - The Gaiko command line interface

USAGE:
   gaiko [global options] command [command options]

VERSION:
   1.14.11-stable

COMMANDS:
   one-shot        Run state transition once
   one-batch-shot  Run multi states transition once
   aggregate       Run the aggregate process
   bootstrap       Run the bootstrap process
   check           Run the check process
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --secret-dir value  Directory for the secret files (default: /Users/xus/.config/gaiko/secrets)
   --config-dir value  Directory for configuration files (default: /Users/xus/.config/gaiko/config)
   --sgx-type value    Which SGX type? ego or gramine [$SGX_TYPE]
   --help, -h          show help
   --version, -v       print the version

COPYRIGHT:
   Copyright 2025-2025 The Gaiko Authors
```
