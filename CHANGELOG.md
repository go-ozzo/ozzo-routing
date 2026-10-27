# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Releases before v2.4.1 are described on the [GitHub releases page](https://github.com/go-ozzo/ozzo-routing/releases).

## [Unreleased]

### BREAKING

- **`auth` uses `github.com/golang-jwt/jwt/v5`** instead of `github.com/golang-jwt/jwt` (v3, unmaintained, advisory
  GO-2025-3553 / CVE-2025-30204 without a fixed version). `JWTTokenHandler`, `DefaultJWTTokenHandler` and `NewJWT`
  take jwt/v5 types, and the token stored under `"JWT"` is a jwt/v5 `*jwt.Token`. Change the import to
  `github.com/golang-jwt/jwt/v5` wherever you handle these tokens. Left on v3, custom token handlers and `NewJWT`
  calls fail to compile, and **`c.Get("JWT").(*jwt.Token)` compiles but panics at runtime**. Validation keeps the
  same rules (signing method pinned; `exp`, `nbf` and `iat` checked); jwt/v5 is stricter at the `exp` second, on
  fractional `exp` values and on trailing data in the claims, and words the 401 messages differently. A v3
  `jwt.TimeFunc` used for clock skew is silently ignored now: use `JWTOptions.Leeway`. See [UPGRADE.md](UPGRADE.md).

### Added

- `file.ServerOptions.FS`: serve files from an `fs.FS`, such as an `embed.FS`, instead of `RootPath`. Names are
  cleaned the same way `http.Dir` cleans them, and the `PathMap` target bounds requests as it does for `RootPath`.
  (#68, @zekroTJA)
- `auth.JWTOptions.Leeway`: the clock skew allowed when checking `exp`, `nbf` and `iat`.

### Changed

- **Go 1.21 or newer is required** (was 1.13). It is the minimum of jwt/v5 and of the `io/fs` support in this
  release, and the oldest Go the CI tests.
- deps: github.com/golang-jwt/jwt v3.2.2 → github.com/golang-jwt/jwt/v5 v5.3.1 (see BREAKING). The v2.4.1
  pre-check for bearer tokens without exactly two periods is removed: jwt/v5 splits tokens without the per-period
  allocation, and a test keeps such headers cheap.
- deps: github.com/stretchr/testify v1.4.0 → v1.12.1.
- deps: github.com/golang/gddo is removed. The repository is archived; `content.LanguageNegotiator` now parses
  `Accept-Language` with a copy of its `ParseAccept` in an internal package, with gddo's tests. Behaviour is unchanged.

### Fixed

- `cors.Handler` answers a preflight with only the requested headers that `AllowHeaders` allows. It used to echo the
  whole `Access-Control-Request-Headers` value as soon as one name was allowed, approving the others too. With an
  explicit `AllowHeaders` list the allowed names now come back comma-joined without the original spacing
  (`X-Ping,X-Pong`); with `AllowHeaders: "*"` the request value is echoed as before. (#72, @bensynapse)

### Internal

- CI on GitHub Actions: tests with the race detector on Linux (oldstable, stable), Windows and macOS (stable) and on
  the minimum Go from `go.mod`, a linux/386 build, golangci-lint v2.15.0, gofmt and `go mod tidy` checks,
  govulncheck, and coverage on Codecov. Travis CI is removed.
- golangci-lint configuration and fixes for its findings; no behaviour changes.
- SECURITY.md, CONTRIBUTING.md, CODEOWNERS, Dependabot for GitHub Actions, `.gitattributes` keeping LF line endings.

## [2.4.1] - 2026-10-27

### Security

- `file.Server` no longer serves files outside the directory a `PathMap` entry maps to. Dot segments in the
  request path, plain or percent-encoded, reached any file under `RootPath`, and a prefix without a trailing slash
  mapped to a target with one (`"/static": "/ui/"`) let `/static../name` reach the directory above the mapped one.
  Such requests now get 404; every other request is resolved exactly as before. ([GHSA-g7pf-jc4r-mx95](https://github.com/go-ozzo/ozzo-routing/security/advisories/GHSA-g7pf-jc4r-mx95))
- `file.Server` and `file.Content` no longer put the server's file path into 404 messages, which the router and
  `fault.ErrorHandler` send to the client. The message is now `Not Found`.
- `auth.JWT` rejects a bearer token that does not contain exactly two periods before handing it to golang-jwt v3,
  whose parser allocates about 16 bytes per byte of such a header
  ([GO-2025-3553](https://pkg.go.dev/vuln/GO-2025-3553), CVE-2025-30204). Valid tokens and the 401 response are
  unchanged. golang-jwt v3 is replaced with jwt/v5 in v2.5.0.

### Fixed

- `file.Server` serves `IndexFile` for directory requests on Windows, where it answered 404 because the index path
  was joined with backslashes.

[Unreleased]: https://github.com/go-ozzo/ozzo-routing/compare/v2.4.1...HEAD
[2.4.1]: https://github.com/go-ozzo/ozzo-routing/compare/v2.4.0...v2.4.1
