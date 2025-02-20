# Gaiko is golang port of Raiko

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
