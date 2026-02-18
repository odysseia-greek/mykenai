# Archimedes

Archimedes is the CLI toolbox for the `mykenai` monorepo. It handles text ingestion, parsing utilities, image builds, scaffold generation, and legacy Odysseia install/docs automation.

εὕρηκα ("I found it")

## Status

`odysseia`-based install flow is currently deprecated and is not used anymore for installs.
It will be restored in the future.

## How Archimedes Fits In Mykenai

This directory contains:

- `cmd/archimedes.go`: CLI entrypoint.
- `command/text`: fetch/crawl Greek + translation passages from Scaife into `rhema-*.json`.
- `command/parse`: list/quiz reshaping helpers for word datasets.
- `command/images`: Docker image build helpers for one service or an entire repo.
- `command/skene`: API/job scaffolding from embedded templates.
- `command/odysseia`: legacy Odysseia install/docs helpers (deprecated install path).
- `eratosthenes/bibliotheke`: source text files and parsing references.

## Build

From `archimedes/`:

```bash
make build
```

`make build` defaults to `build-auto`.

- `make build-auto`: reads the installed Archimedes version (for example `1.0.8`) and builds with patch+1 (for example `1.0.9`).
- `make build-custom VERSION=2.0.0`: builds with the exact version you provide, useful for minor/major bumps.

All build targets move the binary to `~/bin/archimedes`.

## Command Groups

```bash
archimedes --help
```

Top-level command groups:

- `text`: crawl/fetch Scaife passages and write Rhema JSON.
- `parse`: transform and regroup word lists.
- `images`: build/push container images from `Dockerfile`/`Containerfile`.
- `skene`: scaffold new API/job services from templates.
- `odysseia`: legacy install/docs tooling.

## Text Commands

`text` writes data in the Rhema shape (`author`, `book`, `reference`, `rhemai[]`).

### Crawl

Use crawl to walk through passages and write one `rhema-<reference>.json` per section group.

```bash
text crawl --author Herodotus --book Histories --grc urn:cts:greekLit:tlg0016.tlg001.perseus-grc2 --eng urn:cts:greekLit:tlg0016.tlg001.perseus-eng2 --out-dir . --max 3
```

### Fetch

Use fetch to get a single passage or a ranged passage and write one output JSON.

```bash
text fetch --author Herodotus --book Histories --grc 'urn:cts:greekLit:tlg0016.tlg001.perseus-grc2:1.1.0-1.1.4' --eng 'urn:cts:greekLit:tlg0016.tlg001.perseus-eng2:1.1.0-1.1.4' --out-dir .
```

If you run the root binary directly, prepend `archimedes`:

```bash
archimedes text fetch --author Herodotus --book Histories --grc 'urn:cts:greekLit:tlg0016.tlg001.perseus-grc2:1.1.0-1.1.4' --eng 'urn:cts:greekLit:tlg0016.tlg001.perseus-eng2:1.1.0-1.1.4' --out-dir .
```

## Other Useful Commands

### Parse

- `archimedes parse list --filepath <txt> --outdir <dir> --mode by-file|by-letter`
- `archimedes parse reparse --sullego <path> [--all]`
- `archimedes parse group --sullego <path>`
- `archimedes parse one-from-many --sullego <path>`

### Images

- `archimedes images single --root <service-dir> --tag <tag> --target prod`
- `archimedes images repo --repo <repo-dir> --tag <tag> --target prod`

Optional: `--multi` for multi-arch (`linux/arm64,linux/amd64`) buildx builds.

### Skene

- `archimedes skene api --name <service> --repo <repo> --index <index> --port 5000`
- `archimedes skene job --name <service> --repo <repo> --index <index> --embed <dir>`

## External Tooling Used By Commands

Depending on the command group, Archimedes shells out to:

- `docker` / `docker buildx`
- `kubectl`
- `helmfile`
- `npx spectaql`
- `go` (for scaffold initialization)

Make sure these are available in your PATH for the workflows you use.
