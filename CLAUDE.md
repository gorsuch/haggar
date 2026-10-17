# CLAUDE.md

## What this is

haggar is a small load generator for carbon/Graphite. It simulates a swarm of collectd-like agents: every `-spawn-interval` (± `-jitter`), it starts one more agent until it reaches `-agents`. Each agent sends `-metrics` random values over TCP to `-carbon` every `-flush-interval`, using the plaintext line format (`name value epoch\n`). Sending `SIGUSR1` pauses or resumes spawning.

## Layout

It's a single `package main` with no third-party dependencies (standard library only).

- `main.go`: flags (set in `init()`), the `Agent` type with its flush loop, and the spawn loop in `main()`
- `util.go`: pure helpers `genMetricNames` and `carbonate` (writes one carbon line)
- `util_test.go`: example-style tests (`// Output:` blocks)

Config is held in package-level vars populated by `flag`. Agents run as goroutines that never stop. The process runs until it is killed.

## Commands

The Go toolchain is pinned in `mise.toml`. Run `mise install` to set it up.

```sh
mise run test       # go vet + go test (same as CI)
mise run build      # wraps make build
make build          # bin/haggar-$GOOS-$GOARCH (CGO disabled, -trimpath)
make build-linux    # linux/amd64 binary, needed by the Dockerfile
make build-docker   # builds gorsuch/haggar:latest from a scratch image
go vet ./...
go test ./...
```

`bin/` is gitignored.

CI is GitHub Actions (`.github/workflows/test.yml`). It installs Go through mise, then runs `mise run test` and `mise run build` on pushes to `main` and on pull requests.

Example tests for unexported functions must be named `Example_name`, not `ExampleName`, or vet fails.

## Known issues

- The README's `go get` install instruction is outdated. Use `go install github.com/gorsuch/haggar@latest`.
- The README flag list omits `-cache_connections`. Note the underscore: the other flags use hyphens.

## Conventions

- The default branch is `main`.
- Log lines are prefixed with `master:` for the spawn loop and `agent N:` for agents. This "master" means the process role, not the git branch.
- Keep it dependency-free and stdlib-only unless there's a strong reason not to.
