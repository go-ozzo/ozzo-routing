// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cors provides a handler for handling CORS.
package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-ozzo/ozzo-routing"
)

const (
	headerOrigin         = "Origin"
	headerRequestMethod  = "Access-Control-Request-Method"
	headerRequestHeaders = "Access-Control-Request-Headers"

	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowCredentials = "Access-Control-Allow-Credentials"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerExposeHeaders    = "Access-Control-Expose-Headers"
	headerMaxAge           = "Access-Control-Max-Age"
)

// Options specifies how the CORS handler should respond with appropriate CORS headers.
type Options struct {
	// the allowed origins (separated by commas). Use an asterisk (*) to indicate allowing all origins.
	AllowOrigin string
	// whether to allow sending auth credentials such as cookies
	AllowCredentials bool
	// the allowed HTTP methods (separated by commas)
	AllowMethods string
	// the allowed HTTP headers in the request (separated by commas).
	// If not set, it defaults to DefaultAllowHeaders.
	AllowHeaders string
	// the HTTP headers that may be read from the response.
	ExposeHeaders string
	// Max amount of seconds that CORS headers may be cached by the browser.
	MaxAge time.Duration

	allowOriginMap map[string]bool
	allowMethodMap map[string]bool
	allowHeaderMap map[string]bool
}

// DefaultAllowHeaders gives the default allowed HTTP headers when Options.AllowHeaders is not set
var DefaultAllowHeaders = "Origin,Accept,Content-Type,Authorization"

// Handlers creates a routing handler that adds appropriate CORS headers according to the specified options and the request.
func Handler(opts Options) routing.Handler {

	opts.init()

	return func(c *routing.Context) error {
		origin := c.Request.Header.Get(headerOrigin)
		method := c.Request.Header.Get(headerRequestMethod)
		headers := c.Request.Header.Get(headerRequestHeaders)

		if c.Request.Method == "OPTIONS" && (method != "" || headers != "") {
			// a preflight request
			opts.setPreflightHeaders(origin, method, headers, c.Response.Header())
		} else {
			opts.setHeaders(origin, c.Response.Header())
		}
		return nil
	}
}

func (o *Options) init() {
	if o.AllowHeaders == "" {
		o.AllowHeaders = DefaultAllowHeaders
	}
	o.allowHeaderMap = buildAllowMap(o.AllowHeaders)
	o.allowMethodMap = buildAllowMap(o.AllowMethods)
	o.allowOriginMap = buildAllowMap(o.AllowOrigin)
}

func buildAllowMap(s string) map[string]bool {
	m := make(map[string]bool)
	if len(s) > 0 {
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(p)
			m[strings.ToUpper(p)] = true
		}
	}
	return m
}

func (o *Options) isOriginAllowed(origin string) bool {
	return o.AllowOrigin == "*" || o.allowOriginMap[strings.ToUpper(origin)]
}

func (o *Options) setHeaders(origin string, headers http.Header) {
	if !o.isOriginAllowed(origin) {
		return
	}

	o.setCommonHeaders(origin, headers)

	if len(o.AllowMethods) > 0 {
		headers.Set(headerAllowMethods, o.AllowMethods)
	}

	if len(o.AllowHeaders) > 0 {
		headers.Set(headerAllowHeaders, o.AllowHeaders)
	}
}

func (o *Options) setPreflightHeaders(origin, method, reqHeaders string, headers http.Header) {
	if !o.isOriginAllowed(origin) {
		return
	}

	o.setCommonHeaders(origin, headers)

	if o.allowMethodMap[strings.ToUpper(method)] {
		headers.Set(headerAllowMethods, o.AllowMethods)
	}

	var allowed []string
	for _, header := range strings.Split(reqHeaders, ",") {
		header = strings.TrimSpace(header)
		if o.allowHeaderMap[strings.ToUpper(header)] {
			allowed = append(allowed, header)
		}
	}
	if len(allowed) > 0 {
		headers.Set(headerAllowHeaders, strings.Join(allowed, ","))
	}
}

func (o *Options) setCommonHeaders(origin string, headers http.Header) {
	if o.AllowOrigin == "*" {
		headers.Set(headerAllowOrigin, "*")
	} else {
		headers.Set(headerAllowOrigin, origin)
	}

	headers.Set(headerAllowCredentials, strconv.FormatBool(o.AllowCredentials))

	if len(o.ExposeHeaders) > 0 {
		headers.Set(headerExposeHeaders, o.ExposeHeaders)
	}

	if o.MaxAge > time.Duration(0) {
		headers.Set(headerMaxAge, strconv.FormatInt(int64(o.MaxAge/time.Second), 10))
	}
}
