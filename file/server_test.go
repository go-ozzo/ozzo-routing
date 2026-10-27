// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package file

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-ozzo/ozzo-routing/v2"
	"github.com/stretchr/testify/assert"
)

func TestParsePathMap(t *testing.T) {
	tests := []struct {
		id       string
		pathMap  PathMap
		from, to string
	}{
		{"t1", PathMap{}, "", ""},
		{"t2", PathMap{"/": ""}, "/", ""},
		{"t3", PathMap{"/": "ui/dist"}, "/", "ui/dist"},
		{"t4", PathMap{"/abc/123": "ui123/abc", "/abc": "/ui/abc", "/abc/xyz": "/xyzui/abc"}, "/abc,/abc/123,/abc/xyz", "/ui/abc,ui123/abc,/xyzui/abc"},
	}
	for _, test := range tests {
		af, at := parsePathMap(test.pathMap)
		assert.Equal(t, test.from, strings.Join(af, ","), test.id)
		assert.Equal(t, test.to, strings.Join(at, ","), test.id)
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		id        string
		from, to  []string
		url, path string
		found     bool
	}{
		{"t1", []string{}, []string{}, "", "", false},

		{"t2.1", []string{"/"}, []string{"/www"}, "", "", false},
		{"t2.2", []string{"/"}, []string{"/www"}, "/", "/www", true},
		{"t2.3", []string{"/"}, []string{"/www"}, "/index", "/wwwindex", true},
		{"t2.4", []string{"/"}, []string{"/www/"}, "/index", "/www/index", true},
		{"t2.5", []string{"/"}, []string{"/www/"}, "/index/", "/www/index/", true},
		{"t2.6", []string{"/"}, []string{"/www/"}, "index", "", false},
		{"t2.7", []string{""}, []string{""}, "/", "/", true},
		{"t2.7", []string{""}, []string{""}, "/index.html", "/index.html", true},

		{"t3.1", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "", "", false},
		{"t3.2", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/", "/www/others", true},
		{"t3.3", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/css", "/www/ui/css", true},
		{"t3.4", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/abc", "/www/othersabc", true},
		{"t3.5", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/css2", "/www/ui/css2", true},

		{"t4.1", []string{"/css/abc", "/css"}, []string{"/www/abc", "/www/css"}, "/css/abc", "/www/css/abc", true},

		// a URL whose dot segments climb above the mapped target matches nothing
		{"t5.1", []string{"/"}, []string{"/www/"}, "/../x", "", false},
		{"t5.2", []string{"/"}, []string{"/www/"}, "/a/../../x", "", false},
		{"t5.3", []string{"/css"}, []string{"/www/css"}, "/css/../../etc/passwd", "", false},
		{"t5.4", []string{"/"}, []string{"/www/"}, "/..", "", false},
		{"t5.5", []string{"/static"}, []string{"/www/"}, "/static../x", "", false},
		{"t5.6", []string{"/"}, []string{"/www/."}, "/./../w2/x", "", false},
		{"t5.7", []string{"/"}, []string{"/www"}, "//../x", "", false},
		{"t5.14", []string{"/static"}, []string{"/www/css"}, "/static/../css-private/x", "", false},
		{"t5.15", []string{"/static"}, []string{"/www/css"}, "/static-a/../../x", "", false},
		{"t5.16", []string{"/static"}, []string{"/www/css"}, "/static-a/../b/x", "", false},
		{"t5.17", []string{"/"}, []string{""}, "/a/../../x", "a/../../x", true},
		{"t5.18", []string{"/static"}, []string{"/www/css"}, "/static-a/./b", "/www/css-a/./b", true},
		// dot segments that stay under the target are passed on unchanged, as before
		{"t5.8", []string{"/"}, []string{"/www/"}, "/a/./b/../c/", "/www/a/./b/../c/", true},
		{"t5.9", []string{"/"}, []string{"/www/"}, "/a/..", "/www/a/..", true},
		{"t5.10", []string{"/"}, []string{"/www"}, "//./x", "/www/./x", true},
		{"t5.13", []string{"/"}, []string{"/www"}, "/../x", "/www../x", true},
		{"t5.11", []string{"/a"}, []string{"/w/a"}, "/a./a", "/w/a./a", true},
		{"t5.12", []string{"/"}, []string{""}, "/../x", "../x", true},
	}
	for _, test := range tests {
		path, found := matchPath(test.url, test.from, test.to)
		assert.Equal(t, test.found, found, test.id)
		if found {
			assert.Equal(t, test.path, path, test.id)
		}
	}
}

func TestContent(t *testing.T) {
	h := Content("testdata/index.html")
	req, _ := http.NewRequest("GET", "/index.html", nil)
	res := httptest.NewRecorder()
	c := routing.NewContext(res, req)
	err := h(c)
	assert.Nil(t, err)
	assert.Equal(t, "hello\n", res.Body.String())

	h = Content("testdata/index.html")
	req, _ = http.NewRequest("POST", "/index.html", nil)
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, http.StatusMethodNotAllowed, err.(routing.HTTPError).StatusCode())
	}

	h = Content("testdata/index.go")
	req, _ = http.NewRequest("GET", "/index.html", nil)
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, http.StatusNotFound, err.(routing.HTTPError).StatusCode())
	}

	h = Content("testdata/css")
	req, _ = http.NewRequest("GET", "/index.html", nil)
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, http.StatusNotFound, err.(routing.HTTPError).StatusCode())
	}
}

func TestServer(t *testing.T) {
	h := Server(PathMap{"/css": "/testdata/css"})
	tests := []struct {
		id          string
		method, url string
		status      int
		body        string
	}{
		{"t1", "GET", "/css/main.css", 0, "body {}\n"},
		{"t2", "HEAD", "/css/main.css", 0, ""},
		{"t3", "GET", "/css/main2.css", http.StatusNotFound, ""},
		{"t4", "POST", "/css/main.css", http.StatusMethodNotAllowed, ""},
		{"t5", "GET", "/css", http.StatusNotFound, ""},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.url, nil)
		res := httptest.NewRecorder()
		c := routing.NewContext(res, req)
		err := h(c)
		if test.status == 0 {
			assert.Nil(t, err, test.id)
			assert.Equal(t, test.body, res.Body.String(), test.id)
		} else if assert.NotNil(t, err, test.id) {
			assert.Equal(t, test.status, err.(routing.HTTPError).StatusCode(), test.id)
		}
	}

	h = Server(PathMap{"/css": "/testdata/css"}, ServerOptions{
		IndexFile: "index.html",
		Allow: func(c *routing.Context, path string) bool {
			return path != "/testdata/css/main.css"
		},
	})

	req, _ := http.NewRequest("GET", "/css/main.css", nil)
	res := httptest.NewRecorder()
	c := routing.NewContext(res, req)
	err := h(c)
	assert.NotNil(t, err)

	req, _ = http.NewRequest("GET", "/css", nil)
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "css.html\n", res.Body.String())

	{
		// with CatchAll option
		h = Server(PathMap{"/css": "/testdata/css"}, ServerOptions{
			IndexFile:    "index.html",
			CatchAllFile: "testdata/index.html",
			Allow: func(c *routing.Context, path string) bool {
				return path != "/testdata/css/main.css"
			},
		})

		req, _ := http.NewRequest("GET", "/css/main.css", nil)
		res := httptest.NewRecorder()
		c := routing.NewContext(res, req)
		err := h(c)
		assert.NotNil(t, err)

		req, _ = http.NewRequest("GET", "/css", nil)
		res = httptest.NewRecorder()
		c = routing.NewContext(res, req)
		err = h(c)
		assert.Nil(t, err)
		assert.Equal(t, "css.html\n", res.Body.String())

		req, _ = http.NewRequest("GET", "/css2", nil)
		res = httptest.NewRecorder()
		c = routing.NewContext(res, req)
		err = h(c)
		assert.Nil(t, err)
		assert.Equal(t, "hello\n", res.Body.String())
	}
}

type serverCase struct {
	id          string
	method, url string
	status      int
	body        string
}

func runServerCases(t *testing.T, h routing.Handler, tests []serverCase) {
	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.url, nil)
		res := httptest.NewRecorder()
		c := routing.NewContext(res, req)
		err := h(c)
		if test.status == 0 {
			if assert.Nil(t, err, test.id) {
				assert.Equal(t, test.body, res.Body.String(), test.id)
			}
		} else if assert.NotNil(t, err, test.id) {
			assert.Equal(t, test.status, err.(routing.HTTPError).StatusCode(), test.id)
		}
	}
}

// The mapped directory is a boundary: dot segments, plain or percent-encoded, must not
// reach files that are inside RootPath but outside the directory the URL prefix maps to.
func TestServerStaysInMappedDir(t *testing.T) {
	root := Server(PathMap{"/": "/testdata/css/"})
	prefixed := Server(PathMap{"/static": "/testdata/css/"})
	named := Server(PathMap{"/static": "/testdata/css"})
	catchAll := Server(PathMap{"/": "/testdata/css/"}, ServerOptions{CatchAllFile: "testdata/css/main.css"})
	tests := []struct {
		id     string
		h      routing.Handler
		url    string
		status int
		body   string
	}{
		{"inside", root, "/main.css", 0, "body {}\n"},
		{"dot inside", root, "/./main.css", 0, "body {}\n"},
		{"dot dot inside", root, "/sub/../main.css", 0, "body {}\n"},
		{"parent", root, "/../index.html", http.StatusNotFound, ""},
		{"encoded parent", root, "/%2e%2e/index.html", http.StatusNotFound, ""},
		{"source file", root, "/../../server.go", http.StatusNotFound, ""},
		{"encoded source file", root, "/%2e%2e/%2e%2e/server.go", http.StatusNotFound, ""},
		{"encoded slashes", root, "/..%2f..%2fserver.go", http.StatusNotFound, ""},
		{"glued to prefix", prefixed, "/static../index.html", http.StatusNotFound, ""},
		{"glued to prefix, source file", prefixed, "/static../../server.go", http.StatusNotFound, ""},
		{"named target", named, "/static/main.css", 0, "body {}\n"},
		{"named target, sibling with the same prefix", named, "/static/../css-private/secret.txt", http.StatusNotFound, ""},
		{"named target, encoded sibling", named, "/static/%2e%2e/css-private/secret.txt", http.StatusNotFound, ""},
		{"catch all is not a way out", catchAll, "/../../server.go", http.StatusNotFound, ""},
	}
	for _, test := range tests {
		h := test.h
		req, _ := http.NewRequest("GET", test.url, nil)
		res := httptest.NewRecorder()
		err := h(routing.NewContext(res, req))
		if test.status == 0 {
			if assert.Nil(t, err, test.id) {
				assert.Equal(t, test.body, res.Body.String(), test.id)
			}
		} else if assert.NotNil(t, err, test.id) {
			assert.Equal(t, test.status, err.(routing.HTTPError).StatusCode(), test.id)
		}
	}
}

// http.Dir cleans the requested name, so directory URLs with a trailing slash and
// "//", "." or ".." segments that stay inside RootPath are served.
func TestServerCleansNames(t *testing.T) {
	h := Server(PathMap{"/": "/"}, ServerOptions{RootPath: "testdata", IndexFile: "index.html"})
	runServerCases(t, h, []serverCase{
		{"root", "GET", "/", 0, "hello\n"},
		{"dir", "GET", "/css", 0, "css.html\n"},
		{"dir with slash", "GET", "/css/", 0, "css.html\n"},
		{"double slash", "GET", "/css//main.css", 0, "body {}\n"},
		{"dot", "GET", "/./css/main.css", 0, "body {}\n"},
		{"dot dot", "GET", "/css/../index.html", 0, "hello\n"},
		{"outside root", "GET", "/../server.go", http.StatusNotFound, ""},
		{"encoded outside root", "GET", "/%2e%2e/server.go", http.StatusNotFound, ""},
		{"backslash outside root", "GET", "/..%5cserver.go", http.StatusNotFound, ""},
		{"nul", "GET", "/index.html%00", http.StatusNotFound, ""},
	})
}

// An FS backed by a directory must not serve files outside that directory either.
func TestServerFSContainment(t *testing.T) {
	h := Server(PathMap{"/": "/"}, ServerOptions{FS: os.DirFS("testdata")})
	runServerCases(t, h, []serverCase{
		{"inside", "GET", "/index.html", 0, "hello\n"},
		{"outside root", "GET", "/../server.go", http.StatusNotFound, ""},
		{"encoded outside root", "GET", "/%2e%2e/server.go", http.StatusNotFound, ""},
		{"backslash outside root", "GET", "/..%5cserver.go", http.StatusNotFound, ""},
		{"nul", "GET", "/index.html%00", http.StatusNotFound, ""},
	})
}

func TestServerFS(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":     {Data: []byte("fs index\n")},
		"css/index.html": {Data: []byte("fs css index\n")},
		"css/main.css":   {Data: []byte("fs css\n")},
	}

	// FS takes precedence over RootPath: testdata holds files with the same names.
	h := Server(PathMap{"/": "/"}, ServerOptions{FS: fsys, RootPath: "testdata", IndexFile: "index.html"})
	runServerCases(t, h, []serverCase{
		{"file", "GET", "/css/main.css", 0, "fs css\n"},
		{"head", "HEAD", "/css/main.css", 0, ""},
		{"root", "GET", "/", 0, "fs index\n"},
		{"dir", "GET", "/css", 0, "fs css index\n"},
		{"dir with slash", "GET", "/css/", 0, "fs css index\n"},
		{"double slash", "GET", "/css//main.css", 0, "fs css\n"},
		{"dot", "GET", "/./css/main.css", 0, "fs css\n"},
		{"dot dot", "GET", "/css/../index.html", 0, "fs index\n"},
		{"outside root", "GET", "/../index.html", 0, "fs index\n"},
		{"missing", "GET", "/css/main2.css", http.StatusNotFound, ""},
		{"post", "POST", "/css/main.css", http.StatusMethodNotAllowed, ""},
	})

	h = Server(PathMap{"/static": ""}, ServerOptions{FS: fsys})
	runServerCases(t, h, []serverCase{
		{"empty target", "GET", "/static/css/main.css", 0, "fs css\n"},
		{"dir without index file", "GET", "/static/css", http.StatusNotFound, ""},
	})

	h = Server(PathMap{"/": "/"}, ServerOptions{FS: fsys, CatchAllFile: "index.html"})
	runServerCases(t, h, []serverCase{
		{"catch all", "GET", "/missing.css", 0, "fs index\n"},
	})

	// The README example: an embed.FS keeps the embedded directory name in its paths.
	embedded := fstest.MapFS{"ui/index.html": {Data: []byte("ui index\n")}}
	h = Server(PathMap{"/": "/ui/"}, ServerOptions{FS: embedded, IndexFile: "index.html"})
	runServerCases(t, h, []serverCase{
		{"embedded root", "GET", "/", 0, "ui index\n"},
		{"embedded file", "GET", "/index.html", 0, "ui index\n"},
	})
}

// The PathMap target bounds requests served from an FS the same way as from RootPath.
func TestServerFSStaysInMappedDir(t *testing.T) {
	fsys := fstest.MapFS{
		"ui/index.html":             {Data: []byte("ui index\n")},
		"ui/css/main.css":           {Data: []byte("ui css\n")},
		"ui/css-private/secret.txt": {Data: []byte("secret\n")},
	}
	root := Server(PathMap{"/": "/ui/css/"}, ServerOptions{FS: fsys})
	runServerCases(t, root, []serverCase{
		{"inside", "GET", "/main.css", 0, "ui css\n"},
		{"dot dot inside", "GET", "/x/../main.css", 0, "ui css\n"},
		{"parent", "GET", "/../index.html", http.StatusNotFound, ""},
		{"encoded parent", "GET", "/%2e%2e/index.html", http.StatusNotFound, ""},
	})
	prefixed := Server(PathMap{"/static": "/ui/css/"}, ServerOptions{FS: fsys})
	runServerCases(t, prefixed, []serverCase{
		{"glued to prefix", "GET", "/static../index.html", http.StatusNotFound, ""},
	})
	named := Server(PathMap{"/static": "/ui/css"}, ServerOptions{FS: fsys})
	runServerCases(t, named, []serverCase{
		{"named target", "GET", "/static/main.css", 0, "ui css\n"},
		{"named target, sibling with the same prefix", "GET", "/static/../css-private/secret.txt", http.StatusNotFound, ""},
	})
}

// The error message is written to the client, so it must not carry the server's file paths.
func TestNotFoundHidesFilePaths(t *testing.T) {
	handlers := map[string]routing.Handler{
		"server missing":   Server(PathMap{"/": "/testdata/"}),
		"server catch all": Server(PathMap{"/": "/testdata/"}, ServerOptions{CatchAllFile: "missing.html"}),
		"content missing":  Content("testdata/missing.html"),
	}
	for id, h := range handlers {
		req, _ := http.NewRequest("GET", "/missing.css", nil)
		err := h(routing.NewContext(httptest.NewRecorder(), req))
		if assert.NotNil(t, err, id) {
			assert.Equal(t, http.StatusNotFound, err.(routing.HTTPError).StatusCode(), id)
			assert.Equal(t, http.StatusText(http.StatusNotFound), err.Error(), id)
		}
	}
}
