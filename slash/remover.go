// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package slash provides a trailing slash remover handler for the ozzo routing package.
package slash

import (
	"net/http"
	"strings"

	"github.com/go-ozzo/ozzo-routing"
)

// Remover returns a handler that removes the trailing slash (if any) from the requested URL.
// The handler will redirect the browser to the new URL without the trailing slash.
// The status parameter should be either http.StatusMovedPermanently (301) or http.StatusFound (302), which is to
// be used for redirecting GET requests. For other requests, the status code will be http.StatusTemporaryRedirect (307).
// If the original URL has no trailing slash, the handler will do nothing. For example,
//
//     import (
//         "net/http"
//         "github.com/go-ozzo/ozzo-routing"
//         "github.com/go-ozzo/ozzo-routing/slash"
//     )
//
//     r := routing.New()
//     r.Use(slash.Remover(http.StatusMovedPermanently))
//
// Note that Remover relies on HTTP redirection to remove the trailing slashes.
// If you do not want redirection, please set `Router.IgnoreTrailingSlash` to be true without using Remover.
func Remover(status int) routing.Handler {
	return func(c *routing.Context) error {
		if c.Request.URL.Path != "/" && strings.HasSuffix(c.Request.URL.Path, "/") {
			if c.Request.Method != "GET" {
				status = http.StatusTemporaryRedirect
			}
			http.Redirect(c.Response, c.Request, strings.TrimRight(c.Request.URL.Path, "/"), status)
			c.Abort()
		}
		return nil
	}
}
