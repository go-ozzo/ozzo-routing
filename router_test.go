// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"strings"
	"fmt"
	"sort"
)


func TestNewRouter(t *testing.T) {
	var r *Router

	r = NewRouter()
	if r.Pattern != "" {
		t.Errorf("New().pattern = %q, want %q", r.Pattern, "")
	}
	if len(r.Handlers) != 0 {
		t.Errorf("len(New().handlers) = %v, want 0", len(r.Handlers))
	}

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

		// methods must be in upper case
		{"get /users", nil, "get /users"},
	}

	for _, tt := range tests {
		r = NewChildRouter(tt.path, nil)
		if r.Pattern != tt.pattern {
			t.Errorf("NewRouter(%q).pattern = %q, want %q", tt.path, r.Pattern, tt.pattern)
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
			t.Errorf("NewRouter(%q).method = %q, want %q", tt.path, strings.Join(methods, ","), strings.Join(tt.methods, ","))
		}
	}

	// testing handlers
	h1 := func(*Context) {}
	h2 := func(*Context) {}
	r = NewChildRouter("", []Handler{h1, h2})
	if len(r.Handlers) != 2 || r.Handlers[0] == nil || r.Handlers[1] == nil {
		t.Errorf("NewRouter(\"\", h1, h2): handlers not assigned correctly")
	}
}

func TestRouterMatch(t *testing.T) {
	tests := []struct {
		// input
		pattern   string
		method    string
		path      string
		// output
		matching  bool
		remaining string
	}{
		{"", "GET", "", true, ""},
		{"", "GET", "/", true, "/"},

		{"/users", "GET", "/users", true, ""},
		{"/users", "GET", "/user", false, "/user"},
		{"/users", "GET", "/users/123", true, "/123"},
		{"/users", "POST", "/users", true, ""},

		{"GET /users", "GET", "/users", true, ""},
		{"GET /users", "GET", "/user", false, "/user"},
		{"GET /users", "GET", "/users/123", true, "/123"},
		{"GET /users", "POST", "/users", false, "/users"},

		{"GET,POST /users", "GET", "/users", true, ""},
		{"GET,POST /users", "POST", "/users", true, ""},
		{"GET,POST /users", "PATCH", "/users", false, "/users"},

		// regexp
		{"/users/\\d+", "GET", "/users/123", true, ""},
		{"/users/\\d+", "GET", "/users/12a", true, "a"},
		{"/users/\\d+/[a-z]*", "GET", "/users/12/", true, ""},
		{"/users/\\d+/[a-z]*", "GET", "/users/12/abc", true, ""},
		{"/users/\\d+/[a-z]*", "GET", "/users/12/abc1", true, "1"},
	}

	for _, tt := range tests {
		matching, remaining, _ := NewChildRouter(tt.pattern, nil).Match(tt.method, tt.path)
		if matching != tt.matching {
			t.Errorf("NewRouter(%q).Match(%q, %q).matching = %v, want %v", tt.pattern, tt.method, tt.path, matching, tt.matching)
		}
		if remaining != tt.remaining {
			t.Errorf("NewRouter(%q).Match(%q, %q).remaining = %q, want %q", tt.pattern, tt.method, tt.path, remaining, tt.remaining)
		}
	}
}

func TestRouterMatchParams(t *testing.T) {
	tests := []struct {
		// input
		pattern  string
		path     string
		// output
		matching bool
		params   map[string]string
	}{
		{"", "", true, map[string]string{}},
		{"", "/", true, map[string]string{}},

		{"/users", "/users", true, map[string]string{}},
		{"/users", "/user", false, map[string]string{}},

		{"/users/<id>", "/users/", false, map[string]string{}},
		{"/users/<id>", "/users/123abc/", true, map[string]string{"id": "123abc"}},
		{"/users/<id>", "/users/123abc", true, map[string]string{"id": "123abc"}},

		{"/users/<id:\\d+>/<name>", "/users/123", false, map[string]string{}},
		{"/users/<id:\\d+>/<name>", "/users/123/", false, map[string]string{}},
		{"/users/<id:\\d+>/<name>", "/users/123/abc", true, map[string]string{"id": "123", "name": "abc"}},
		{"/users/<id:\\d+>/<name>", "/users/123/abc/", true, map[string]string{"id": "123", "name": "abc"}},
	}

	for _, tt := range tests {
		matching, _, params := NewChildRouter(tt.pattern, nil).Match("GET", tt.path)
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
			t.Errorf("NewRouter(%q).Match(%q, %q).params = %v, want %v", tt.pattern, "GET", tt.path, params, tt.params)
		}
	}
}

func TestDispatchBasic(t *testing.T) {
	r := NewRouter()
	r.Get("", handle("/"))
	r.Get("/users", handle("users/index"))
	r.Get("/users/123", handle("users/view"))
	r.Post("/users/123", handle("users/create"))
	r.To("HEAD,PATCH /users/123", handle("users/any"))

	tests := []dispatchTest{
		{"GET", "", "</>"},
		{"GET", "/", ""},

		{"GET", "/users", "<users/index>"},
		{"GET", "/users/", ""},
		{"POST", "/users", ""},

		{"GET", "/users/123", "<users/view>"},
		{"POST", "/users/123", "<users/create>"},

		{"PATCH", "/users/123", "<users/any>"},
		{"HEAD", "/users/123", "<users/any>"},
		{"OPTION", "/users/123", ""},

		{"GET", "/posts", ""},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchMethods(t *testing.T) {
	r := NewRouter()
	r.Get("/users", handle("users/get"))
	r.Put("/users", handle("users/put"))
	r.Post("/users", handle("users/post"))
	r.Patch("/users", handle("users/patch"))
	r.Delete("/users", handle("users/delete"))
	r.Head("/users", handle("users/head"))
	r.Use(handleNext("posts/use3"))
	r.Options("/users", handle("users/options"))
	r.Use(handleNext("posts/use1"), handle("posts/use2"))

	tests := []dispatchTest{
		{"GET", "/users", "<users/get>"},
		{"PUT", "/users", "<users/put>"},
		{"POST", "/users", "<users/post>"},
		{"PATCH", "/users", "<users/patch>"},
		{"DELETE", "/users", "<users/delete>"},
		{"HEAD", "/users", "<users/head>"},
		{"OPTIONS", "/users", "<posts/use3<users/options>posts/use3>"},
		{"OPTIONS", "/posts", "<posts/use3<posts/use1<posts/use2>posts/use1>posts/use3>"},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchMultiple(t *testing.T) {
	r := NewRouter()
	r.Get("/users", handle("users1"), handle("users2"))
	r.Get("/users", handle("users4"))

	tests := []dispatchTest{
		{"GET", "/users", "<users1>"},
		{"GET", "/users/abc", ""},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchRegexp(t *testing.T) {
	r := NewRouter()
	r.Get("/users", handle("users/index"))
	r.Get("/users/\\d+", handle("users/view"))

	tests := []dispatchTest{
		{"GET", "/users/123", "<users/view>"},
		{"GET", "/users/123/", ""},
		{"GET", "/users/abc", ""},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchNext(t *testing.T) {
	r := NewRouter()
	r.Get("/users", handleNext("users1"), handleNext("users2"))
	r.Get("/users", handleNext("users3"), handle("users4"), handleNext("users5"))
	r.Get("/posts", handleNextRoute("posts1"), handleNext("posts2"))
	r.Get("/posts", handleNext("posts3"), handle("posts4"), handleNext("posts5"))

	tests := []dispatchTest{
		{"GET", "/users", "<users1<users2<users3<users4>users3>users2>users1>"},
		{"GET", "/posts", "<posts1<posts3<posts4>posts3>posts1>"},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchGroup(t *testing.T) {
	r := NewRouter()
	r.Get("/users", handle("users"))
	r.Group("/admin", func(r *Router) {
		r.Get("/users", handleNext("ausers1"), handleNext("ausers2"))
		r.Get("/users", handleNext("ausers3"), handle("ausers4"), handleNext("ausers5"))
		r.Get("/posts", handleNextRoute("aposts1"), handleNext("aposts2"))
		r.Get("/posts", handleNext("aposts3"), handle("aposts4"), handleNext("aposts5"))
		r.Get("/comments", handleNext("acomments"))
		r.Get("/tags", handleNextRoute("atags"))
	})
	r.Get("/admin/comments", handle("comments"))
	r.Get("/admin/tags", handle("tags"))

	tests := []dispatchTest{
		{"GET", "/users", "<users>"},
		{"GET", "/admin", ""},
		{"GET", "/admin/users", "<ausers1<ausers2<ausers3<ausers4>ausers3>ausers2>ausers1>"},
		{"GET", "/admin/posts", "<aposts1<aposts3<aposts4>aposts3>aposts1>"},
		{"GET", "/admin/comments", "<acomments<comments>acomments>"},
		{"GET", "/admin/tags", "<atags<tags>atags>"},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchGroupDeeply(t *testing.T) {
	r := NewRouter()
	r.Get("/users", handle("users"))
	r.Group("/admin", func(r *Router) {
		r.Get("/users", handle("admin/users"))
		r.Group("/profile", func(r *Router) {
			r.Get("/address", handle("address"))
			r.Get("/posts", handleNextRoute("posts3"))
		})
		r.Get("/profile/posts", handleNext("posts2"))
	})
	r.Get("/admin/profile/posts", handle("posts1"))

	tests := []dispatchTest{
		{"GET", "/users", "<users>"},
		{"GET", "/admin", ""},
		{"GET", "/admin/users", "<admin/users>"},
		{"GET", "/admin/users/profile", ""},
		{"GET", "/admin/profile/address", "<address>"},
		{"GET", "/admin/profile/posts", "<posts3<posts2<posts1>posts2>posts3>"},
	}

	runDispatchTests(t, tests, r)
}

func TestDispatchGroupHandlers(t *testing.T) {
	r := NewRouter()
	r.Group("/admin", func(r *Router) {
		r.Get("/users", handle("users"))
	}, handleNext("admin"))
	r.Get("/admin/users", handle("users2"))
	r.Group("/posts", func(r *Router) {
		r.Get("/comments", handle("comments"))
	}, handleNext("posts1"), handle("posts2"), handle("posts3"))
	r.Group("/tags", func(r *Router) {
		r.Get("/cloud", handle("cloud"))
	}, handleNextRoute("tags1"), handle("tags2"))
	r.Get("/tags/cloud", handle("tags3"))

	tests := []dispatchTest{
		{"GET", "/admin/users", "<admin<users>admin>"},
		{"GET", "/admin/u", "<adminadmin>"},
		{"GET", "/posts/comments", "<posts1<posts2>posts1>"},
		{"GET", "/posts/c", "<posts1<posts2>posts1>"},
		{"GET", "/tags/cloud", "<tags1<tags3>tags1>"},
		{"GET", "/tags/c", "<tags1tags1>"},
	}

	runDispatchTests(t, tests, r)
}

func TestErrorHandling(t *testing.T) {
	r := NewRouter()
	r.Get("/users", triggerError("users", false))
	r.Get("/posts", triggerError("posts1", true), triggerError("posts2", false), triggerError("posts3", false))
	r.Error(handleError("users", true))
	r.Error(handleError("posts1", true))
	r.Error(handleError("posts2", true))
	r.Error(handleError("users1", false))
	r.Error(handleError("users2", true))
	r.Group("/admin", func(r *Router) {
		r.Get("/users", triggerError("ausers", false))
		r.Get("/posts", triggerError("aposts1", true), triggerError("aposts2", false), triggerError("aposts3", false))
		r.Error(handleError("posts2", true))
	})
	r.Error(handleError("posts2a", true))
	r.Error(handleError("ausers", true))
	r.Error(handleError("aposts1", true))

	tests := []struct {
		Error  string
		Path   string
		Result string
	}{
		{"users", "/xyz", ""},
		{"users", "/users", "<users<err:users><err:users1>"},
		{"posts1", "/posts", "<posts1<err:posts1>"},
		{"posts2", "/posts", "<posts1<posts2<err:posts2><err:posts2a>posts1>"},
		{"posts3", "/posts", "<posts1<posts2posts2>posts1>"},
		{"ausers", "/admin/users", "<ausers<err:ausers>"},
		{"aposts1", "/admin/posts", "<aposts1<err:aposts1>"},
	}

	for _, test := range tests {
		errorToken = test.Error
		req, _ := http.NewRequest("GET", test.Path, nil)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Body.String() != test.Result {
			t.Errorf("Error = %q, Dispatch(%q, %q) = %q, want %q", test.Error, "GET", test.Path, res.Body.String(), test.Result)
		}
	}
}

func fmtMap(m map[string]string) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := ""
	for _, k := range keys {
		result += k + ":" + m[k] + ","
	}
	return "{" + result + "}"
}

func TestDispatchParams(t *testing.T) {
	r := NewRouter()
	r.Get("/users/<id:\\d+>/<name>", handle("users"))
	r.Group("/posts/<id:\\d+>", func(r *Router) {
		r.Get("/<name>", handle("posts"))
	})
	r.Group("/comments/<id:\\d+>", func(r *Router) {
		r.Get("/<name>", handleNext("comments1"))
	}, handleNext("comments2"))
	r.Get("/comments/<data:.*>", handle("comments3"))

	tests := []dispatchTest{
		{"GET", "/users", ""},
		{"GET", "/users/", ""},
		{"GET", "/users/123", ""},
		{"GET", "/users/123/", ""},
		{"GET", "/users/123/abc/", ""},
		{"GET", "/users/123/abc", "<users>{id:123,name:abc,}"},
		{"GET", "/posts/123/abc", "<posts>{id:123,name:abc,}"},
		{"GET", "/comments/123/abc", "<comments2{id:123,}<comments1{id:123,name:abc,}<comments3>{data:123/abc,}comments1>comments2>"},
		{"GET", "/comments/123abc", "<comments2{id:123,}<comments3>{data:123abc,}comments2>"},
	}

	runDispatchTests(t, tests, r)
}

var handle = func(token string) Handler {
	return func(c *Context) {
		fmt.Fprint(c.Response, "<" + token + ">")
		if len(c.Params) > 0 {
			fmt.Fprint(c.Response, fmtMap(c.Params))
		}
	}
}

var handleNext = func(token string) Handler {
	return func(c *Context) {
		fmt.Fprint(c.Response, "<" + token)
		if len(c.Params) > 0 {
			fmt.Fprint(c.Response, fmtMap(c.Params))
		}
		c.Next()
		fmt.Fprint(c.Response, token + ">")
	}
}

var handleNextRoute = func(token string) Handler {
	return func(c *Context) {
		fmt.Fprint(c.Response, "<" + token)
		if len(c.Params) > 0 {
			fmt.Fprint(c.Response, fmtMap(c.Params))
		}
		c.NextRoute()
		fmt.Fprint(c.Response, token + ">")
	}
}

var errorToken = ""

var triggerError = func(token string, next bool) Handler {
	return func(c *Context) {
		fmt.Fprint(c.Response, "<" + token)
		if errorToken == token {
			panic(token)
		}
		if next {
			c.Next()
		}
		fmt.Fprint(c.Response, token + ">")
	}
}

var handleError = func(token string, next bool) Handler {
	return func(c *Context) {
		if c.Error != nil && strings.HasPrefix(token, c.Error.(string)) {
			fmt.Fprint(c.Response, "<err:" + token + ">")
			if next {
				c.Next()
			}
		} else {
			c.Next()
		}
	}
}

type dispatchTest struct {
	method string
	path   string
	result string
}

func runDispatchTests(t *testing.T, tests []dispatchTest, r *Router) {
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Body.String() != tt.result {
			t.Errorf("Dispatch(%q, %q) = %q, want %q", tt.method, tt.path, res.Body.String(), tt.result)
		}
	}
}
