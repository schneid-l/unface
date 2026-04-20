# Contributing

Thanks for your interest in unface! This project is maintained in the open.

## Development setup

Requires Go 1.25+ and `golangci-lint` v2+.

```bash
git clone https://github.com/schneid-l/unface
cd unface
go test ./...
golangci-lint run
```

## Conventions

- **Conventional Commits.** Prefixes: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `perf`, `ci`.
- **Atomic commits.** Every commit must build and pass tests + lint independently.
- **Tests first.** Add a failing test before the fix/feature.
- **No TODOs** committed to `main`.

## PR checklist

- [ ] Tests added / updated
- [ ] `go vet ./...` passes
- [ ] `golangci-lint run` passes
- [ ] `go test -race ./...` passes
- [ ] Godoc on new exports
- [ ] `CHANGELOG.md` updated under `[Unreleased]`
