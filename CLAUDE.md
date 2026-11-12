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

## Releases

Current release: **v2.0.0** (2026-11-09). It is breaking: the MySQL and SQLite engines and the `--debug` / `--dbcreds` flags were removed (PR #7). It also includes the code-review fixes (PR #6). Downstream apps (webapp-starter and the apps built from it) use `livewireholdings/dbp:latest` in their `Dockerfile.dbp`. None of them used the removed options.

- Work on a branch → PR → merge (merge commit, not squash). There is no CI on PRs, so run `go vet ./... && go test ./...` before merging. For engine/patcher changes, also apply a real patch set to a throwaway Postgres twice (the second run should print "No work to do").
- `.github/workflows/docker-publish.yaml` runs on **every push to `main`** and on `v*` tags. It runs the tests, then builds amd64+arm64 and pushes `livewireholdings/dbp:latest` plus `:$(make versiontag)`. `docker-verify.yaml` then checks that both platforms are in the manifest.
- To release: annotated tag on `main` (`git tag -a vX.Y.Z -m "…"`), then `git push origin vX.Y.Z`. Semver: removing an engine, flag or behaviour is a major bump.
- Every merge to `main` also moves `:latest`. If several runs overlap, check that `:latest` and the new tag have the same digest afterwards (`docker buildx imagetools inspect livewireholdings/dbp:<tag>`), and re-run the tag's workflow if they don't.

## Known quirks

- `apply --dry` doesn't consult the database, so on an already-patched database it still lists every patch as "applying".
- `testdata/certs` (and `testdata/env/pg/certs`, which links to it) is gitignored and exists only locally, so `testdata/env/pg` `make start` fails in a fresh clone.

## Architecture

`dbp` is a forward-only SQL database patcher CLI. It walks a directory tree of `.sql` patch files, resolves their prerequisite graph, and applies unapplied patches in dependency order — each inside its own transaction alongside an insert into `dbp_patch_table`.

### Package layout

- **`main.go`** — wires `cmd.Version` from build-time ldflags, calls `cmd.Execute()`
- **`cmd/`** — Cobra subcommands: `apply`, `validate`, `new`, `version`, `root`
- **`pkg/patch/`** — `Patch` struct (id, prereqs, options, body, weight) and `ByWeight` sort implementation
- **`pkg/patcher/`** — core logic: `Scan` walks directories, `Resolve` builds the prereq graph and detects loops via recursive `bumpWeight`, `Process` applies patches in order
- **`pkg/dbe/`** — `DBEngine` interface + implementations: `pg_dbe.go`, `mock_dbe.go`; `EngineArgs` reads connection params from CLI flags or environment variables
- **`pkg/set/`** — simple string set used to track installed patch IDs

### How patching works

1. `Scan` walks the folder for `*.sql` files, parsing comment headers (`-- id:`, `-- prereqs:`, `-- options:`) via regex
2. `Resolve` builds the dependency graph: `bumpWeight` recursively increments weights on all prerequisites, detecting cycles with a `detectionMap`; patches are sorted descending by weight (higher weight → applied earlier, so prerequisites come first)
3. `Process` applies uninstalled patches in order; the init patch (`init_patch.sql`) is applied first when no patches have been recorded yet
4. Each patch applies inside a transaction that also inserts the patch ID into `dbp_patch_table` — both succeed or neither does

### Patch file format

```sql
-- id: some-unique-id
-- prereqs: id-of-patch-a,id-of-patch-b
-- description: optional human description
-- options: chop

...SQL here...
```

- Header keys are only read from the comments before the first line of SQL, and the first occurrence of a key wins
- `id` is required; any string without spaces or commas (UUID recommended; `dbp new` generates one)
- `prereqs` is a comma-separated list of IDs that must be applied first
- `options: chop` splits the file into statements before executing — useful for multi-statement files where you want per-statement error messages. A `;` inside a comment, quoted string, quoted identifier or dollar-quoted string doesn't split, and the last statement doesn't need a `;`
- One file named exactly `init_patch.sql` must exist in the tree; it's the first patch applied to a new database (`dbp_patch_table` itself is created by the engine)

### DBEngine interface

```go
type DBEngine interface {
    GetInstalledIDs() (*set.Set, error)
    Patch(*patch.Patch) error
}
```

Postgres is the only supported database, because dbp relies on transactional DDL to roll back a failed patch. `MockDBE` is used in tests and by the `validate` command (dry-run, no real DB needed).

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

`pkg/patcher/testdata/` contains patch hierarchies demonstrating patcher behavior, and `testdata/env/pg` is for manual testing:
- `pkg/patcher/testdata/bad/{dupe_id,long_loop,short_loop,shortest_loop,missing_id_1,missing_id_2}` — error cases
- `pkg/patcher/testdata/good/patchset_1` — a valid patch set
- `testdata/env/pg` — docker-compose environment for manual integration testing

`TestPatcher` runs these cases against `MockDBE`. `TestIgnore` and `TestParseHeader` use patches written to a temp folder.
