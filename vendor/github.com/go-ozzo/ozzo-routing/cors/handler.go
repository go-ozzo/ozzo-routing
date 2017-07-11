// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
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
	headerOrigin = "Origin"

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
	// the allowed origins (separated by commas). Use an asterisk (*) to indicate allowing all origins, "null" to indicate disallowing any.
	AllowOrigins string
	// whether the response to request can be exposed when the omit credentials flag is unset, or whether the actual request can include user credentials.
	AllowCredentials bool
	// the HTTP methods (separated by commas) that can be used during the actual request. Use an asterisk (*) to indicate allowing any method.
	AllowMethods string
	// the HTTP headers (separated by commas) that can be used during the actual request. Use an asterisk (*) to indicate allowing any header.
	AllowHeaders string
	// the HTTP headers (separated by commas) that are safe to expose to the API of a CORS API specification
	ExposeHeaders string
	// Max amount of seconds that the results of a preflight request can be cached in a preflight result cache.
	MaxAge time.Duration

	allowOriginMap map[string]bool
	allowMethodMap map[string]bool
	allowHeaderMap map[string]bool
}

// AllowAll is the option that allows all origins, headers, and methods.
var AllowAll = Options{
	AllowOrigins: "*",
	AllowHeaders: "*",
	AllowMethods: "*",
}

// Handler creates a routing handler that adds appropriate CORS headers according to the specified options and the request.
func Handler(opts Options) routing.Handler {

	opts.init()

	return func(c *routing.Context) (err error) {
		origin := c.Request.Header.Get(headerOrigin)
		if origin == "" {
			// the request is outside the scope of CORS
			return
		}
		if c.Request.Method == "OPTIONS" {
			// a preflight request
			method := c.Request.Header.Get(headerRequestMethod)
			if method == "" {
				// the request is outside the scope of CORS
				return
			}
			headers := c.Request.Header.Get(headerRequestHeaders)
			opts.setPreflightHeaders(origin, method, headers, c.Response.Header())
			c.Abort()
			return
		}
		opts.setActualHeaders(origin, c.Response.Header())
		return
	}
}

func (o *Options) init() {
	o.allowHeaderMap = buildAllowMap(o.AllowHeaders, false)
	o.allowMethodMap = buildAllowMap(o.AllowMethods, true)
	o.allowOriginMap = buildAllowMap(o.AllowOrigins, true)
}

func (o *Options) isOriginAllowed(origin string) bool {
	if o.AllowOrigins == "null" {
		return false
	}
	return o.AllowOrigins == "*" || o.allowOriginMap[origin]
}

func (o *Options) setActualHeaders(origin string, headers http.Header) {
	if !o.isOriginAllowed(origin) {
		return
	}

	o.setOriginHeader(origin, headers)

	if o.ExposeHeaders != "" {
		headers.Set(headerExposeHeaders, o.ExposeHeaders)
	}
}

func (o *Options) setPreflightHeaders(origin, method, reqHeaders string, headers http.Header) {
	allowed, allowedHeaders := o.isPreflightAllowed(origin, method, reqHeaders)
	if !allowed {
		return
	}

	o.setOriginHeader(origin, headers)

	if o.MaxAge > time.Duration(0) {
		headers.Set(headerMaxAge, strconv.FormatInt(int64(o.MaxAge/time.Second), 10))
	}

	if o.AllowMethods == "*" {
		headers.Set(headerAllowMethods, method)
	} else if o.allowMethodMap[method] {
		headers.Set(headerAllowMethods, o.AllowMethods)
	}

	if allowedHeaders != "" {
		headers.Set(headerAllowHeaders, reqHeaders)
	}
}

func (o *Options) isPreflightAllowed(origin, method, reqHeaders string) (allowed bool, allowedHeaders string) {
	if !o.isOriginAllowed(origin) {
		return
	}
	if o.AllowMethods != "*" && !o.allowMethodMap[method] {
		return
	}
	if o.AllowHeaders == "*" || reqHeaders == "" {
		return true, reqHeaders
	}

	headers := []string{}
	for _, header := range strings.Split(reqHeaders, ",") {
		header = strings.TrimSpace(header)
		if o.allowHeaderMap[strings.ToUpper(header)] {
			headers = append(headers, header)
		}
	}
	if len(headers) > 0 {
		return true, strings.Join(headers, ",")
	}
	return
}

func (o *Options) setOriginHeader(origin string, headers http.Header) {
	if o.AllowCredentials {
		headers.Set(headerAllowOrigin, origin)
		headers.Set(headerAllowCredentials, "true")
	} else {
		if o.AllowOrigins == "*" {
			headers.Set(headerAllowOrigin, "*")
		} else {
			headers.Set(headerAllowOrigin, origin)
		}
	}
}

func buildAllowMap(s string, caseSensitive bool) map[string]bool {
	m := make(map[string]bool)
	if len(s) > 0 {
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(p)
			if caseSensitive {
				m[p] = true
			} else {
				m[strings.ToUpper(p)] = true
			}
		}
	}
	return m
}
