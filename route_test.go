// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"testing"
	"strings"
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		// input
		path    string
		// output
		methods []string
		pattern string
	}{
		// normal cases
		{"", nil, ""},
		{"/users", nil, "/users"},
		{"GET /users", []string{"GET"}, "/users"},
		{"GET,POST /users", []string{"GET", "POST"}, "/users"},


		// required space separators between method, pattern and name
		{"GET/users", nil, "GET/users"},
		{"GET /users@us", []string{"GET"}, "/users@us"},

		// methods must be in upper case
		{"get /users", nil, "get /users"},
	}

	var r *Route
	for _, tt := range tests {
		r = NewRoute(tt.path, nil)
		if r.Pattern != tt.pattern {
			t.Errorf("newRoute(%q).pattern = %q, want %q", tt.path, r.Pattern, tt.pattern)
		}
		hasError := len(r.Methods) != len(tt.methods)
		if !hasError {
			for _, m := range tt.methods {
				if !r.Methods[m] {
					hasError = true
					break
				}
			}
		}
		if hasError {
			methods := make([]string, 0)
			for m := range r.Methods {
				methods = append(methods, m)
			}
			t.Errorf("newRoute(%q).method = %q, want %q", tt.path, strings.Join(methods, ","), strings.Join(tt.methods, ","))
		}
	}

	// testing handlers
	h1 := func(Context) {}
	h2 := func(Context) {}
	r = NewRoute("", []Handler{h1, h2})
	if len(r.Handlers) != 2 || r.Handlers[0] == nil || r.Handlers[1] == nil {
		t.Errorf("newRoute(\"\", h1, h2): handlers not assigned correctly")
	}
}

func TestRouteMatch(t *testing.T) {
	tests := []struct {
		// input
		pattern  string
		method   string
		path     string
		// output
		matching bool
	}{
		{"", "GET", "", true},
		{"", "GET", "/", false},

		{"/users", "GET", "/users", true},
		{"/users", "GET", "/user", false},
		{"/users", "GET", "/users/123", false},
		{"/users", "POST", "/users", true},

		{"/users.html", "POST", "/users.html", true},
		{"/users.html", "POST", "/usersahtml", true},
		{"/users\\.html", "POST", "/users.html", true},
		{"/users\\.html", "POST", "/usersahtml", false},

		{"GET /users", "GET", "/users", true},
		{"GET /users", "GET", "/user", false},
		{"GET /users", "GET", "/users/123", false},
		{"GET /users", "POST", "/users", false},

		{"GET,POST /users", "GET", "/users", true},
		{"GET,POST /users", "POST", "/users", true},
		{"GET,POST /users", "PATCH", "/users", false},

		// regexp
		{"/users/\\d+", "GET", "/users/123", true},
		{"/users/\\d+", "GET", "/users/12a", false},
		{"/users/\\d+/[a-z]*", "GET", "/users/12/", true},
		{"/users/\\d+/[a-z]*", "GET", "/users/12/abc", true},
		{"/users/\\d+/[a-z]*", "GET", "/users/12/abc1", false},
	}

	for _, tt := range tests {
		matching, _, _ := NewRoute(tt.pattern, nil).Match(tt.method, tt.path)
		if matching != tt.matching {
			t.Errorf("newRoute(%q).Match(%q, %q) = %v, want %v", tt.pattern, tt.method, tt.path, matching, tt.matching)
		}
	}
}

func TestRouteMatchParams(t *testing.T) {
	tests := []struct {
		// input
		pattern  string
		path     string
		// output
		matching bool
		params   map[string]string
	}{
		{"", "", true, map[string]string{}},
		{"", "/", false, map[string]string{}},

		{"/users", "/users", true, map[string]string{}},
		{"/users", "/user", false, map[string]string{}},

		{"/users/<id>", "/users/", false, map[string]string{}},
		{"/users/<id>", "/users/123abc/", false, map[string]string{}},
		{"/users/<id>", "/users/123abc", true, map[string]string{"id": "123abc"}},

		{"/users/<id:\\d+>/<name>", "/users/123", false, map[string]string{}},
		{"/users/<id:\\d+>/<name>", "/users/123/", false, map[string]string{}},
		{"/users/<id:\\d+>/<name>", "/users/123/abc", true, map[string]string{"id": "123", "name": "abc"}},
		{"/users/<id:\\d+>/<name>", "/users/123/abc/", false, map[string]string{}},
	}

	for _, tt := range tests {
		matching, _, params := NewRoute(tt.pattern, nil).Match("GET", tt.path)
		if matching != tt.matching {
			t.Errorf("newRoute(%q).Match(%q, %q).matching = %v, want %v", tt.pattern, "GET", tt.path, matching, tt.matching)
		}

		hasError := len(params) != len(tt.params)
		if !hasError {
			for k, v := range tt.params {
				if params[k] != v {
					hasError = true
					break
				}
			}
		}
		if hasError {
			t.Errorf("newRoute(%q).Match(%q, %q).params = %v, want %v", tt.pattern, "GET", tt.path, params, tt.params)
		}
	}
}

