// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package auth provides a set of user authentication handlers for the ozzo routing package.
package auth

import (
	"net/http"
	"encoding/base64"
	"github.com/go-ozzo/ozzo-routing"
)

// User is the key used to store and retrieve the user identity information in routing.Context
const User = "User"

// Identity represents an authenticated user. If a user is successfully authenticated by
// an auth handler (Basic, Bearer, or Query), an Identity object will be made available for injection.
type Identity interface{}

// DefaultRealm is the default realm name for HTTP authentication. It is used by HTTP authentication based on
// Basic and Bearer.
var DefaultRealm = "API"

// BasicAuthFunc is the function that does the actual user authentication according to the given username and password.
type BasicAuthFunc func(c *routing.Context, username, password string) (Identity, error)

// Basic returns a routing.Handler that performs HTTP basic authentication.
// It can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/go-ozzo/ozzo-routing"
//     "github.com/go-ozzo/ozzo-routing/auth"
//   )
//   func main() {
//     r := routing.New()
//     r.Use(auth.Basic(func(c *routing.Context, username, password string) (auth.Identity, error) {
//       if username == "demo" && password == "foo" {
//         return auth.Identity(username), nil
//       }
//       return nil, errors.New("invalid credential")
//     })
//     r.Get("/demo", func(c *routing.Context) error {
//       fmt.Fprintf(res, "Hello, %v", c.Get(auth.User))
//       return nil
//     })
//   }
//
// By default, the auth realm is named as "API". You may customize it by specifying the realm parameter.
//
// When authentication fails, a "WWW-Authenticate" header will be sent, and an http.StatusUnauthorized
// error will be reported via routing.Context.Error().
func Basic(fn BasicAuthFunc, realm ...string) routing.Handler {
	name := DefaultRealm
	if len(realm) > 0 {
		name = realm[0]
	}
	return func(c *routing.Context) error {
		username, password, _ := c.Request.BasicAuth()
		identity, e := fn(c, username, password)
		if e == nil {
			c.Set(User, identity)
			return nil
		}
		c.Response.Header().Set("WWW-Authenticate", `Basic realm="` + name + `"`)
		return routing.NewHTTPError(http.StatusUnauthorized, e.Error())
	}
}

// TokenAuthFunc is the function for authenticating a user based on a secret token.
type TokenAuthFunc func(c *routing.Context, token string) (Identity, error)

// Bearer returns a routing.Handler that performs HTTP authentication based on bearer token.
// It can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/go-ozzo/ozzo-routing"
//     "github.com/go-ozzo/ozzo-routing/auth"
//   )
//   func main() {
//     r := routing.New()
//     r.Use(auth.Bearer(func(c *routing.Context, token string) (auth.Identity, error) {
//       if token == "secret" {
//         return auth.Identity("demo"), nil
//       }
//       return nil, errors.New("invalid credential")
//     })
//     r.Get("/demo", func(c *routing.Context) error {
//       fmt.Fprintf(res, "Hello, %v", c.Get(auth.User))
//       return nil
//     })
//   }
//
// By default, the auth realm is named as "API". You may customize it by specifying the realm parameter.
//
// When authentication fails, a "WWW-Authenticate" header will be sent, and an http.StatusUnauthorized
// error will be reported via routing.Context.Error().
func Bearer(fn TokenAuthFunc, realm ...string) routing.Handler {
	name := DefaultRealm
	if len(realm) > 0 {
		name = realm[0]
	}
	return func(c *routing.Context) error {
		token := parseBearerToken(c.Request)
		identity, e := fn(c, token)
		if e == nil {
			c.Set(User, identity)
			return nil
		}
		c.Response.Header().Set("WWW-Authenticate", `Bearer realm="` + name + `"`)
		return routing.NewHTTPError(http.StatusUnauthorized, e.Error())
	}
}

func parseBearerToken(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if len(auth) < 7 || auth[:7] != "Bearer " {
		return ""
	}
	if bearer, err := base64.StdEncoding.DecodeString(auth[7:]); err == nil {
		return string(bearer)
	}
	return ""
}

// TokenName is the query parameter name for auth token.
var TokenName = "access-token"

// Query returns a routing.Handler that performs authentication based on a token passed via a query parameter.
// It can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/go-ozzo/ozzo-routing"
//     "github.com/go-ozzo/ozzo-routing/auth"
//   )
//   func main() {
//     r := routing.New()
//     r.Use(auth.Query(func(token string) (auth.Identity, error) {
//       if token == "secret" {
//         return auth.Identity("demo"), nil
//       }
//       return nil, errors.New("invalid credential")
//     })
//     r.Get("/demo", func(c *routing.Context) error {
//       fmt.Fprintf(res, "Hello, %v", c.Get(auth.User))
//       return nil
//     })
//   }
//
// When authentication fails, an http.StatusUnauthorized error will be reported via routing.Context.Error().
func Query(fn TokenAuthFunc, tokenName ...string) routing.Handler {
	name := TokenName
	if len(tokenName) > 0 {
		name = tokenName[0]
	}
	return func(c *routing.Context) error {
		token := c.Request.URL.Query().Get(name)
		identity, err := fn(c, token)
		if err != nil {
			return routing.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		c.Set(User, identity)
		return nil
	}
}
