# glightning2

## Generate Files

Requirements:

- Go
- Python 3
- A Core Lightning checkout


The generator expects a Core Lightning checkout and an output directory:

```bash
tools/generate-clnrpc.sh <cln-repo-path> <output-path>
```

Example using the local context checkout:

```bash
tools/generate-clnrpc.sh .context/lightning clnrpc
```

The script:

1. Runs `tools/cln-schema-export.py` with Core Lightning's `msggen`.
2. Writes the temporary schema dump to a temporary directory.
3. Runs `go run ./cmd/clnrpcgen`.
4. Removes temporary files automatically.

Only generated Go files remain in the output directory.

