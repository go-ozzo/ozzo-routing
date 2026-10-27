# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Releases before v2.4.1 are described on the [GitHub releases page](https://github.com/go-ozzo/ozzo-routing/releases).

## [Unreleased]

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
