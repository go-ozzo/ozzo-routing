# Contributing to ozzo-routing

Thank you for your interest in contributing! This document covers how to build, test, and submit changes.

## Prerequisites

- **Go 1.21+** ([download](https://go.dev/dl/))
- **golangci-lint v2** ([install](https://golangci-lint.run/welcome/install/)); CI runs v2.15.0

## Building and Testing

```bash
go build ./...
go vet ./...
go test -race ./...
golangci-lint run --timeout=5m
gofmt -l .
```

CI runs the tests on Linux, Windows and macOS and on the minimum Go version from `go.mod`. Path handling in the
`file` package behaves differently on Windows, so a change there needs tests that pass on both.

## Code Style

- Run `gofmt` before every commit. CI enforces it.
- Fix every `golangci-lint` finding, or explain an exception in `.golangci.yml` or a `//nolint` comment with a reason.
- Handle every error, or ignore it explicitly with `_ =` and a comment.
- Exported types and functions must have doc comments.
- The module is `v2`: exported names, signatures and interfaces stay backward compatible.

## Pull Request Workflow

1. Fork the repository and create a feature branch:
   ```bash
   git checkout -b fix/my-fix
   ```
2. Make your changes with tests. A bug fix comes with a test that fails without it.
3. Add a line to the `[Unreleased]` section of `CHANGELOG.md` for any user-visible change.
4. Verify locally with the commands above.
5. Push and open a pull request against `master`. All checks must be green before merge.

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):
```
feat: add fs.FS support to file.Server
fix: filter disallowed headers in CORS preflight responses
docs: update installation instructions
```

## Security Issues

Do not open public issues for vulnerabilities; see [SECURITY.md](SECURITY.md).

## Reporting Bugs

Open an issue with a clear description, a minimal code example that reproduces the problem, and the Go version you're
using.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
