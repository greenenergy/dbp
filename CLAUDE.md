# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build binaries
make                    # builds build/amd64/dbp and build/arm64/dbp
make install            # installs to GOPATH via go install

# Test
make test               # runs all tests with coverage
go test ./pkg/patcher/  # run a single package's tests
go test -run TestIgnore ./pkg/patcher/  # run a single test

# Docker
make docker             # multi-arch push to livewireholdings/dbp
make docker-local       # single-arch amd64 local build

# Version tagging (must use annotated tags)
git tag -a vX.Y.Z
```

Version is injected at build time via `-X main.Version=$(GIT_VERSION)`. The version string strips `-0-ghash` on exact tags and keeps `-ghash` suffix otherwise (see `Makefile` sed expressions).

## Architecture

`dbp` is a forward-only SQL database patcher CLI. It walks a directory tree of `.sql` patch files, resolves their prerequisite graph, and applies unapplied patches in dependency order — each inside its own transaction alongside an insert into `dbp_patch_table`.

### Package layout

- **`main.go`** — wires `cmd.Version` from build-time ldflags, calls `cmd.Execute()`
- **`cmd/`** — Cobra subcommands: `apply`, `validate`, `new`, `version`, `root`
- **`pkg/patch/`** — `Patch` struct (id, prereqs, options, body, weight) and `ByWeight` sort implementation
- **`pkg/patcher/`** — core logic: `Scan` walks directories, `Resolve` builds the prereq graph and detects loops via recursive `bumpWeight`, `Process` applies patches in order
- **`pkg/dbe/`** — `DBEngine` interface + implementations: `pg_dbe.go`, `mysql_dbe.go`, `sqlite_dbe.go`, `mock_dbe.go`; `EngineArgs` reads connection params from CLI flags or environment variables
- **`pkg/set/`** — simple string set used to track installed patch IDs

### How patching works

1. `Scan` walks the folder for `*.sql` files, parsing comment headers (`-- id:`, `-- prereqs:`, `-- options:`) via regex
2. `Resolve` builds the dependency graph: `bumpWeight` recursively increments weights on all prerequisites, detecting cycles with a `detectionMap`; patches are sorted descending by weight (higher weight → applied later)
3. `Process` applies uninstalled patches in order; the init patch (`init_patch.sql`) is applied first when `dbp_patch_table` doesn't yet exist
4. Each patch applies inside a transaction that also inserts the patch ID into `dbp_patch_table` — both succeed or neither does

### Patch file format

```sql
-- id: some-unique-id
-- prereqs: id-of-patch-a,id-of-patch-b
-- description: optional human description
-- options: chop

...SQL here...
```

- `id` is required; any string without spaces or commas (UUID recommended; `dbp new` generates one)
- `prereqs` is a comma-separated list of IDs that must be applied first
- `options: chop` splits the file on `;` before executing — useful for multi-statement files where you want per-statement error messages
- One file named exactly `init_patch.sql` must exist in the tree; it bootstraps `dbp_patch_table`

### DBEngine interface

```go
type DBEngine interface {
    GetInstalledIDs() (*set.Set, error)
    Patch(*patch.Patch) error
}
```

New database engines implement this interface. `MockDBE` is used in tests and by the `validate` command (dry-run, no real DB needed).

### Environment variables for connection params

| Env var | CLI flag | Default |
|---|---|---|
| `DB_HOST` | `--db.host` | |
| `DB_USER` | `--db.username` | |
| `DB_PASSWORD` | `--db.password` | |
| `DB_NAME` | `--db.name` | |
| `DB_SSLMODE` | `--db.sslmode` | `require` |
| `DB_SSLCERT` / `DB_SSLKEY` / `DB_SSLROOTCERT` | corresponding flags | |

CLI flags override env vars. Default SSL mode is `require`.

### Test data

`testdata/` contains patch hierarchies demonstrating patcher behavior:
- `testdata/bad/{dupe_id,long_loop,short_loop,shortest_loop,missing_id_1,missing_id_2}` — error cases
- `testdata/good/patchset_1` — a valid patch set
- `testdata/env/{pg,mysql,sqlite}` — docker-compose environments for manual integration testing

The main `TestPatcher` test is currently skipped (`t.Skip("temp")`). `TestIgnore` runs and exercises the folder-ignore path with `MockDBE`.
