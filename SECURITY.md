# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| >= 2.4.1 | :white_check_mark: |
| < 2.4.1  | :x:                |

Only the module path `github.com/go-ozzo/ozzo-routing/v2` receives fixes. Releases published under
`github.com/go-ozzo/ozzo-routing` (v1.x and v2.0.x–v2.1.x, before the module moved to `/v2`) are not maintained.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report it privately with **Report a vulnerability** on the repository's
[Security tab](https://github.com/go-ozzo/ozzo-routing/security), which opens a private advisory draft:
https://github.com/go-ozzo/ozzo-routing/security/advisories/new

Please include:

- a description of the vulnerability and its impact;
- a minimal reproduction (router setup and the request that triggers it);
- the affected versions you checked.

### Response Timeline

- **Initial response**: within 72 hours
- **Fix and disclosure**: coordinated with the reporter; credit is given unless you prefer otherwise

Published advisories: https://github.com/go-ozzo/ozzo-routing/security/advisories

## Security Considerations

ozzo-routing sits at the edge of an application. Things to keep in mind:

1. **Static files** — `file.Server` does not let `.` and `..` segments in a request climb above the target of the
   `PathMap` entry that matched, or above `ServerOptions.RootPath`. Matching is by string prefix, though: a prefix or
   target without a trailing slash also matches longer names (`"/static": "/www/css"` serves `/static-private/x` from
   `/www/css-private`), so end both with `/` when they name a directory. `http.Dir` follows symbolic links, so keep
   none under `RootPath` that point elsewhere, and set `RootPath` to the directory that holds only the files to serve.
2. **Request paths** — the router matches routes on the request path as received. Put an `http.ServeMux` or a
   path-normalizing proxy in front of it if route-level middleware (authentication, for example) must also apply to
   paths with `.`, `..` or `//` segments.
3. **CORS** — `cors.AllowAll` allows every origin; with `AllowCredentials` list the trusted origins instead.
4. **JWT** — `auth.JWT` accepts only the configured signing method; keep `JWTOptions.SigningMethod` set explicitly
   and the verification key out of the source code.
5. **Error messages** — `fault.ErrorHandler` and the router write error messages to the client; do not put internal
   details into errors returned from handlers.
