# glightning2 - next-generation SDK for the [C-Lightning](https://github.com/ElementsProject/lightning).

The repository contains generated models according to the official schemas and JSON RPC for calling C-Lightning daemon
from your Go application. Note that the repository is WIP:

- [x] Type definitions
- [x] RPC client
- [ ] Plugin builder
- [ ] CLN-rest client
- [ ] Commando Client CLI

### Repository structure

- [clntypes](./clntypes) contains Go representations of Core Lightning schema-specific scalar types.

- [clnrpc](./clnrpc) contains generated Go structs for the current Core Lightning schema model, a JSON-RPC
  client for the native `lightning-rpc` Unix socket, and typed methods for every
  generated RPC.

- [tools](./tools) contains a Python generator that reuses Core Lightning's own `msggen` package directly.
  It loads the normalized CLN service model, applies CLN's `msggen` patches, and
  emits Go structs and typed client methods.

### Update generated structs

To regenerate Go structs for Lightning schema model execute:

```bash
tools/generate-clnrpc.sh <cln-repo-path> <output-path>
```

It requires an up-to-date version of [ElementsProject/lightning](https://github.com/ElementsProject/lightning)
repository to be cloned.

Example:

```bash
tools/generate-clnrpc.sh ~/GolandProjects/lightning clnrpc
```

The wrapper sets `PYTHONPATH` to `<cln-repo-path>/contrib/msggen`, so `msggen` does not need to be installed globally.

## Usage

### RPC Client

```go
client := clnrpc.NewClient("~/.lighting/regtest/lightning-rpc")
info, err := client.GetInfo(ctx)
funds, err := client.ListFunds(ctx, clnrpc.ListFundsRequest{})
```




