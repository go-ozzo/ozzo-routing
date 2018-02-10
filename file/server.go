// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package file provides handlers that serve static files for the ozzo routing package.
package file

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-ozzo/ozzo-routing"
)

// ServerOptions defines the possible options for the Server handler.
type ServerOptions struct {
	// The path that all files to be served should be located within. The path map passed to the Server method
	// are all relative to this path. This property can be specified as an absolute file path or a path relative
	// to the current working path. If not set, this property defaults to the current working path.
	RootPath string
	// The file (e.g. index.html) to be served when the current request corresponds to a directory.
	// If not set, the handler will return a 404 HTTP error when the request corresponds to a directory.
	// This should only be a file name without the directory part.
	IndexFile string
	// The file to be served when no file or directory matches the current request.
	// If not set, the handler will return a 404 HTTP error when no file/directory matches the request.
	// The path of this file is relative to RootPath
	CatchAllFile string
	// A function that checks if the requested file path is allowed. If allowed, the function
	// may do additional work such as setting Expires HTTP header.
	// The function should return a boolean indicating whether the file should be served or not.
	// If false, a 404 HTTP error will be returned by the handler.
	Allow func(*routing.Context, string) bool
}

// PathMap specifies the mapping between URL paths (keys) and file paths (keys).
// The file paths are relative to Options.RootPath
type PathMap map[string]string

// RootPath stores the current working path
var RootPath string

func init() {
	RootPath, _ = os.Getwd()
}

// Server returns a handler that serves the files as the response content.
// The files being served are determined using the current URL path and the specified path map.
// For example, if the path map is {"/css": "/www/css", "/js": "/www/js"} and the current URL path
// "/css/main.css", the file "<working dir>/www/css/main.css" will be served.
// If a URL path matches multiple prefixes in the path map, the most specific prefix will take precedence.
// For example, if the path map contains both "/css" and "/css/img", and the URL path is "/css/img/logo.gif",
// then the path mapped by "/css/img" will be used.
//
//     import (
//         "log"
//         "github.com/go-ozzo/ozzo-routing"
//         "github.com/go-ozzo/ozzo-routing/file"
//     )
//
//     r := routing.New()
//     r.Get("/*", file.Server(file.PathMap{
//          "/css": "/ui/dist/css",
//          "/js": "/ui/dist/js",
//     }))
func Server(pathMap PathMap, opts ...ServerOptions) routing.Handler {
	var options ServerOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	if !filepath.IsAbs(options.RootPath) {
		options.RootPath = filepath.Join(RootPath, options.RootPath)
	}
	from, to := parsePathMap(pathMap)

	// security measure: limit the files within options.RootPath
	dir := http.Dir(options.RootPath)

	return func(c *routing.Context) error {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			return routing.NewHTTPError(http.StatusMethodNotAllowed)
		}
		path, found := matchPath(c.Request.URL.Path, from, to)
		if !found || options.Allow != nil && !options.Allow(c, path) {
			return routing.NewHTTPError(http.StatusNotFound)
		}

		var (
			file  http.File
			fstat os.FileInfo
			err   error
		)

		if file, err = dir.Open(path); err != nil {
			if options.CatchAllFile != "" {
				return serveFile(c, dir, options.CatchAllFile)
			}
			return routing.NewHTTPError(http.StatusNotFound, err.Error())
		}
		defer file.Close()

		if fstat, err = file.Stat(); err != nil {
			return routing.NewHTTPError(http.StatusNotFound, err.Error())
		}

		if fstat.IsDir() {
			if options.IndexFile == "" {
				return routing.NewHTTPError(http.StatusNotFound)
			}
			return serveFile(c, dir, filepath.Join(path, options.IndexFile))
		}

		c.Response.Header().Del("Content-Type")
		http.ServeContent(c.Response, c.Request, path, fstat.ModTime(), file)
		return nil
	}
}

func serveFile(c *routing.Context, dir http.Dir, path string) error {
	file, err := dir.Open(path)
	if err != nil {
		return routing.NewHTTPError(http.StatusNotFound, err.Error())
	}
	defer file.Close()
	fstat, err := file.Stat()
	if err != nil {
		return routing.NewHTTPError(http.StatusNotFound, err.Error())
	} else if fstat.IsDir() {
		return routing.NewHTTPError(http.StatusNotFound)
	}
	c.Response.Header().Del("Content-Type")
	http.ServeContent(c.Response, c.Request, path, fstat.ModTime(), file)
	return nil
}

// Content returns a handler that serves the content of the specified file as the response.
// The file to be served can be specified as an absolute file path or a path relative to RootPath (which
// defaults to the current working path).
// If the specified file does not exist, the handler will pass the control to the next available handler.
func Content(path string) routing.Handler {
	if !filepath.IsAbs(path) {
		path = filepath.Join(RootPath, path)
	}
	return func(c *routing.Context) error {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			return routing.NewHTTPError(http.StatusMethodNotAllowed)
		}
		file, err := os.Open(path)
		if err != nil {
			return routing.NewHTTPError(http.StatusNotFound, err.Error())
		}
		defer file.Close()
		fstat, err := file.Stat()
		if err != nil {
			return routing.NewHTTPError(http.StatusNotFound, err.Error())
		} else if fstat.IsDir() {
			return routing.NewHTTPError(http.StatusNotFound)
		}
		c.Response.Header().Del("Content-Type")
		http.ServeContent(c.Response, c.Request, path, fstat.ModTime(), file)
		return nil
	}
}

func parsePathMap(pathMap PathMap) (from, to []string) {
	from = make([]string, len(pathMap))
	to = make([]string, len(pathMap))
	n := 0
	for i := range pathMap {
		from[n] = i
		n++
	}
	sort.Strings(from)
	for i, s := range from {
		to[i] = pathMap[s]
	}
	return
}

func matchPath(path string, from, to []string) (string, bool) {
	for i := len(from) - 1; i >= 0; i-- {
		prefix := from[i]
		if strings.HasPrefix(path, prefix) {
			return to[i] + path[len(prefix):], true
		}
	}
	return "", false
}
