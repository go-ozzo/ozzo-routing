// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import "strings"

// RouteGroup represents a group of routes that share the same path prefix.
type RouteGroup struct {
	prefix   string
	router   *Router
	handlers []Handler
}

// newRouteGroup creates a new RouteGroup with the given path prefix, router, and handlers.
func newRouteGroup(prefix string, router *Router, handlers []Handler) *RouteGroup {
	return &RouteGroup{
		prefix:   prefix,
		router:   router,
		handlers: handlers,
	}
}

// Get adds a GET route to the router with the given route path and handlers.
func (rg *RouteGroup) Get(path string, handlers ...Handler) *Route {
	return rg.add("GET", path, handlers)
}

// Post adds a POST route to the router with the given route path and handlers.
func (rg *RouteGroup) Post(path string, handlers ...Handler) *Route {
	return rg.add("POST", path, handlers)
}

// Put adds a PUT route to the router with the given route path and handlers.
func (rg *RouteGroup) Put(path string, handlers ...Handler) *Route {
	return rg.add("PUT", path, handlers)
}

// Patch adds a PATCH route to the router with the given route path and handlers.
func (rg *RouteGroup) Patch(path string, handlers ...Handler) *Route {
	return rg.add("PATCH", path, handlers)
}

// Delete adds a DELETE route to the router with the given route path and handlers.
func (rg *RouteGroup) Delete(path string, handlers ...Handler) *Route {
	return rg.add("DELETE", path, handlers)
}

// Connect adds a CONNECT route to the router with the given route path and handlers.
func (rg *RouteGroup) Connect(path string, handlers ...Handler) *Route {
	return rg.add("CONNECT", path, handlers)
}

// Head adds a HEAD route to the router with the given route path and handlers.
func (rg *RouteGroup) Head(path string, handlers ...Handler) *Route {
	return rg.add("HEAD", path, handlers)
}

// Options adds an OPTIONS route to the router with the given route path and handlers.
func (rg *RouteGroup) Options(path string, handlers ...Handler) *Route {
	return rg.add("OPTIONS", path, handlers)
}

// Trace adds a TRACE route to the router with the given route path and handlers.
func (rg *RouteGroup) Trace(path string, handlers ...Handler) *Route {
	return rg.add("TRACE", path, handlers)
}

// Any adds a route with the given route, handlers, and the HTTP methods as listed in routing.Methods.
func (rg *RouteGroup) Any(path string, handlers ...Handler) *Route {
	return rg.To(strings.Join(Methods, ","), path, handlers...)
}

// To adds a route to the router with the given HTTP methods, route path, and handlers.
// Multiple HTTP methods should be separated by commas (without any surrounding spaces).
func (rg *RouteGroup) To(methods, path string, handlers ...Handler) *Route {
	mm := strings.Split(methods, ",")
	if len(mm) == 1 {
		return rg.add(methods, path, handlers)
	}

	r := rg.newRoute(methods, path)
	for _, method := range mm {
		r.routes = append(r.routes, rg.add(method, path, handlers))
	}
	return r
}

// Group creates a RouteGroup with the given route path prefix and handlers.
// The new group will combine the existing path prefix with the new one.
// If no handler is provided, the new group will inherit the handlers registered
// with the current group.
func (rg *RouteGroup) Group(prefix string, handlers ...Handler) *RouteGroup {
	if len(handlers) == 0 {
		handlers = make([]Handler, len(rg.handlers))
		copy(handlers, rg.handlers)
	}
	return newRouteGroup(rg.prefix+prefix, rg.router, handlers)
}

// Use registers one or multiple handlers to the current route group.
// These handlers will be shared by all routes belong to this group and its subgroups.
func (rg *RouteGroup) Use(handlers ...Handler) {
	rg.handlers = append(rg.handlers, handlers...)
}

func (rg *RouteGroup) add(method, path string, handlers []Handler) *Route {
	r := rg.newRoute(method, path)
	rg.router.addRoute(r, combineHandlers(rg.handlers, handlers))
	return r
}

// newRoute creates a new Route with the given route path and route group.
func (rg *RouteGroup) newRoute(method, path string) *Route {
	return &Route{
		group:    rg,
		method:   method,
		path:     path,
		template: buildURLTemplate(rg.prefix + path),
	}
}

// combineHandlers merges two lists of handlers into a new list.
func combineHandlers(h1 []Handler, h2 []Handler) []Handler {
	hh := make([]Handler, len(h1)+len(h2))
	copy(hh, h1)
	copy(hh[len(h1):], h2)
	return hh
}

// buildURLTemplate converts a route pattern into a URL template by removing regular expressions in parameter tokens.
func buildURLTemplate(path string) string {
	path = strings.TrimRight(path, "*")
	template, start, end := "", -1, -1
	for i := 0; i < len(path); i++ {
		if path[i] == '<' && start < 0 {
			start = i
		} else if path[i] == '>' && start >= 0 {
			name := path[start+1 : i]
			for j := start + 1; j < i; j++ {
				if path[j] == ':' {
					name = path[start+1 : j]
					break
				}
			}
			template += path[end+1:start] + "<" + name + ">"
			end = i
			start = -1
		}
	}
	if end < 0 {
		template = path
	} else if end < len(path)-1 {
		template += path[end+1:]
	}
	return template
}
