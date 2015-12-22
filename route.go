// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"regexp"
	"fmt"
	"strings"
)

// Route is a route associated with a list of handlers.
// If a route matches the current HTTP request, the associated handlers will be invoked.
// A route matches a request only if it matches both the HTTP method and the URL path of the current request.
type Route struct {
	Methods  map[string]bool // HTTP methods
	Pattern  string          // URL path to be matched
	Handlers []Handler       // handlers associated with this route

	err      bool            // whether this route is for handling errors
	regex    *regexp.Regexp  // parsed regex of pattern
}

// RoutePatternError describes the route pattern which is of invalid format.
type RoutePatternError string

// Error returns the error message represented by RoutePatternError
func (s RoutePatternError) Error() string {
	return "Invalid route pattern: " + string(s)
}

var (
	routeRegex = regexp.MustCompile(`^(?:([A-Z\,]+)\s+)?(.*?)$`)
	literalRegex = regexp.MustCompile(`^[\w\-~]*$`)
	paramRegex = regexp.MustCompile(`<([^>]+)>`)
	paramInternalRegex = regexp.MustCompile(`^(\w+):?([^>]+)?$`)
)

// NewRoute creates a new route with the specified URL pattern and handlers.
// The URL pattern should be in the format of "METHOD pattern", where "METHOD"
// is optional, representing one or multiple HTTP methods separated by commas,
// while "pattern" is a regular expression used to determine if the route matches
// the currently requested URL path.
//
// The followings are some examples of the pattern parameter:
//
//     /users                // matches "/users"
//     /users/<id:\d+>       // matches "/users/123"
//     GET,POST /users       // matches "/users" for GET or POST only
func NewRoute(pattern string, handlers []Handler) *Route {
	matches := routeRegex.FindStringSubmatch(pattern)
	if len(matches) != 3 {
		panic(RoutePatternError(pattern))
	}

	route := Route{
		Methods: make(map[string]bool),
		Pattern: matches[2],
	}

	if len(matches[1]) > 0 {
		for _, method := range strings.Split(matches[1], ",") {
			route.Methods[method] = true
		}
	}

	route.Handlers = append(route.Handlers, handlers...)

	if !literalRegex.MatchString(route.Pattern) {
		route.regex = regexp.MustCompile("^" + parseParamPattern(route.Pattern) + "$")
	}

	return &route
}

// Match checks if the route matches the specified HTTP method and URL path.
func (r *Route) Match(method, path string) (bool, string, map[string]string) {
	if len(r.Methods) > 0 && !r.Methods[method] {
		return false, path, nil
	}
	return r.MatchPath(path)
}

// MatchPath checks if the route matches the specified URL path
func (r *Route) MatchPath(path string) (bool, string, map[string]string) {
	if r.regex == nil {
		return path == r.Pattern, path, nil
	}

	if r.Pattern == ".*" {
		return true, path, nil
	}

	matches := r.regex.FindStringSubmatch(path)
	if len(matches) == 0 || matches[0] != path {
		return false, path, nil
	}

	params := make(map[string]string)
	for i, name := range r.regex.SubexpNames() {
		if len(name) > 0 {
			params[name] = matches[i]
		}
	}

	return true, path, params
}

// Dispatch invokes the handlers associated with this route.
func (r *Route) Dispatch(method, path string, c *Context) {
	index := 0
	oldNext := c.Next

	c.Next = func() {
		if index < len(r.Handlers) && (r.err == (c.Error != nil)) {
			handler := r.Handlers[index]
			index++
			callHandler(c, handler)
		} else {
			index = len(r.Handlers)
			c.Next = oldNext
			oldNext()
		}
	}

	c.Next()
}

// parseParamPattern converts "<name:pattern>" tokens in the pattern into named subpattern in a regexp.
func parseParamPattern(pattern string) string {
	return paramRegex.ReplaceAllStringFunc(pattern, func(m string) string {
		matches := paramInternalRegex.FindStringSubmatch(m[1 : len(m) - 1])
		switch {
		case len(matches) < 3:
			return m
		case matches[2] == "":
			return fmt.Sprintf(`(?P<%s>[^/]+)`, matches[1])
		default:
			return fmt.Sprintf(`(?P<%s>%s)`, matches[1], matches[2])
		}
	})
}
