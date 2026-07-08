# glightning2

## What Is Here

### `clntypes`

Go representations of Core Lightning schema-specific scalar types, such as
`msat`, `sat`, `hash`, `secret`, `txid`, `outpoint`, `pubkey`,
`short_channel_id`, `short_channel_id_dir`, `feerate`, and `outputdesc`.

### `tools`

A Python generator that reuses Core Lightning's own `msggen` package directly.
It loads the normalized CLN service model, applies CLN's `msggen` patches, and
emits Go structs.

### `clnrpc`

Generated Go structs for the current Core Lightning schema model.

## Howto Generate Files

```bash
tools/generate-clnrpc.sh <cln-repo-path> <output-path>
```

Example:

```bash
tools/generate-clnrpc.sh .context/lightning clnrpc
```

The wrapper sets `PYTHONPATH` to `<cln-repo-path>/contrib/msggen`, so `msggen`
does not need to be installed globally.

