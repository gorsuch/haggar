# CLAUDE.md

## What this is

haggar is a small load generator for carbon/Graphite. It simulates a swarm of collectd-like agents: every `-spawn-interval` (± `-jitter`), it starts one more agent until it reaches `-agents`. Each agent sends `-metrics` random values over TCP to `-carbon` every `-flush-interval`, using the plaintext line format (`name value epoch\n`). Sending `SIGUSR1` pauses or resumes spawning.

## Layout

It's a single `package main` with no third-party dependencies (standard library only).

- `main.go`: flags (set in `init()`), the `Agent` type with its flush loop, and `spawnLoop` (spawn timing, jitter, `SIGUSR1` pause/resume). `main()` just wires `spawnLoop` to the real signal channel and `launchAgent`.
- `util.go`: pure helpers `genMetricNames` and `carbonate` (writes one carbon line)
- `util_test.go`: example-style tests (`// Output:` blocks)
- `agent_test.go`: `Agent.flush` tests against a local TCP listener, plus `BenchmarkAgentFlush`
- `main_test.go`: `spawnLoop` tests. They run under `testing/synctest`, so timers use a fake clock and finish instantly. Use the same approach for any new timing-dependent tests.

Config is held in package-level vars populated by `flag`. Agents run as goroutines that never stop. The process runs until it is killed.

## Commands

The Go toolchain is pinned in `mise.toml`. Run `mise install` to set it up.

```sh
mise run test       # go vet + go test (same as CI)
mise run build      # wraps make build
make build          # bin/haggar-$GOOS-$GOARCH (CGO disabled, -trimpath)
make build-linux    # linux/amd64 binary
make build-docker   # gorsuch/haggar:latest for the local platform (compiles inside Docker)
make push-docker    # multi-arch (linux/amd64, linux/arm64) build and push via buildx
go vet ./...
go test ./...
```

`bin/` is gitignored.

CI is GitHub Actions (`.github/workflows/test.yml`). It installs Go through mise, then runs `mise run test` and `mise run build` on pushes to `main` and on pull requests. A separate `docker` job builds the image for linux/amd64 and linux/arm64 without pushing.

Example tests for unexported functions must be named `Example_name`, not `ExampleName`, or vet fails.

## Conventions

- Flags use hyphens (`-flush-interval`). The exception is `-cache_connections`; keep it for compatibility.
- When adding or changing a flag, update the `haggar -h` output in the README.
- When upgrading Go, update `mise.toml`, the `golang:` image tag in the `Dockerfile`, and possibly the `go` line in `go.mod`.
- The default branch is `main`.
- Log lines are prefixed with `master:` for the spawn loop and `agent N:` for agents. This "master" means the process role, not the git branch.
- Keep it dependency-free and stdlib-only unless there's a strong reason not to.
