# Upgrade Guide

## v2.4 → v2.5: `auth` uses golang-jwt/jwt/v5

`auth.JWT`, `auth.NewJWT`, `auth.JWTTokenHandler` and `auth.DefaultJWTTokenHandler` now use
`github.com/golang-jwt/jwt/v5` instead of `github.com/golang-jwt/jwt` (v3). golang-jwt v3 is no longer maintained
and has an advisory without a fixed version (GO-2025-3553, CVE-2025-30204). Applications that do not use the `auth`
JWT functions are not affected.

### What to change

Replace the import in every file that touches tokens from `auth`:

```diff
-import "github.com/golang-jwt/jwt"
+import "github.com/golang-jwt/jwt/v5"
```

If you still import the older `github.com/dgrijalva/jwt-go`, replace that one the same way. Then run `go mod tidy`.

Most code needs nothing else, for example a token handler or reading the claims:

```go
func handleToken(c *routing.Context, token *jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)
	c.Set(auth.User, claims["id"])
	return nil
}

claims := c.Get("JWT").(*jwt.Token).Claims.(jwt.MapClaims)
```

### What breaks if the import is left on v3

- **Compile errors (loud):** a custom `JWTOptions.TokenHandler` declared with the v3 `*jwt.Token`, and
  `auth.NewJWT(jwt.MapClaims{...}, key)` or a v3 `jwt.SigningMethod` passed to it.
- **Runtime panic (silent until a request comes in):** `c.Get("JWT").(*jwt.Token)` with the v3 import still
  compiles, because `c.Get` returns `interface{}`, but panics with an interface conversion error on the first
  authenticated request. Search your code for `c.Get("JWT")` and check its import.

### Behaviour

- The rules are the same: only the configured signing method is accepted, and `exp`, `nbf` and `iat` are checked
  (a token issued in the future is rejected); `exp: 0` still means "not set". Failed authentication still answers
  401 with `WWW-Authenticate`.
- jwt/v5 is stricter at the edges, for correctly signed tokens only:
  - a token is rejected from the start of its `exp` second; v3 accepted it until the end of that second;
  - a non-zero `exp` between -1 and 1 (`0.5`) is an expiry time, not "not set" as in v3;
  - a claims segment with data after the JSON object is rejected.
- **Clock skew.** v3 let you allow skew by replacing the global `jwt.TimeFunc`. jwt/v5 has no such global, and a
  `jwt.TimeFunc` set through a leftover v3 import is silently ignored. Use the new `JWTOptions.Leeway` instead:

  ```go
  r.Use(auth.JWT(signingKey, auth.JWTOptions{Leeway: 30 * time.Second}))
  ```

  The same applies to `jwt.RegisterSigningMethod` on v3: register custom methods with jwt/v5.
- The 401 message comes from jwt/v5 and is worded differently, for example `token has invalid claims: token is
  expired` instead of `Token is expired`, and `token signature is invalid: signature is invalid` instead of
  `signature is invalid`. Do not match on these texts.
- jwt/v5 API changes that matter only if you use them directly: `Token.Signature` is `[]byte` (was the encoded
  string), `StandardClaims` is gone (use `RegisteredClaims`), claims validate through `jwt.Validator` instead of a
  `Valid()` method. See the [jwt/v5 migration guide](https://github.com/golang-jwt/jwt/blob/main/MIGRATION_GUIDE.md).
