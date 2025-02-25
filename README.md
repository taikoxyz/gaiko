# Gaiko is Golang port of Raiko

## Architecture

```
.
├── cmd
│   └── gaiko
├── ego                 // ego artifacts
└── internal
    ├── mpt             // golang porting mpt implementation from raiko
    ├── transition      // state transition for some blocks
    ├── types           // types compatible layer between gaiko/raiko
    └── version         // binary version helper

```

## Command Help

```
$ ./gaiko
Found existing alias for "go run". You should use: "gor"
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
