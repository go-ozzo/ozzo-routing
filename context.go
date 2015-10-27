// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"net/http"
	"github.com/go-ozzo/ozzo-di"
)

// Context represents the contextual data and environment while processing an incoming HTTP request.
//
// Context is injected into a handler if the handler requires a *Context parameter.
//
// Context contains references to http.Request and http.ResponseWriter which are commonly
// needed by handlers. Context also provides the the URL parameter values through its Params field.
//
// Context serves as a dependency injection container. Handlers may register typed data which can then
// be injected as parameter values for other handlers. By default, if a handler has a parameter of type
// "*routing.Context", the current Context will be injected.
//
// Handlers can use the Data field to share data among them.
//
// When a handler panics, the error will be recovered and made accessible through the Error field.
//
// Within a handler, you may call Context.Next() to pass the control to the next eligible handler;
// call Context.NextRoute() to pass the control to the first handler of the next matching route.
type Context struct {
	di.Container                     // dependency injection container

	Request   *http.Request          // the current HTTP request
	Response  http.ResponseWriter    // the response writer
	Params    map[string]string      // the URL parameter values of the matching route(s)
	Data      map[string]interface{} // the data shared by applicable handlers
	Error     interface{}            // the error recovered from panic

	Next      func()                 // Next invokes the next handler on the current route
	NextRoute func()                 // NextRoute invokes the first handler on the next matching route
}

// NewContext creates a new Context with the given response and request information.
func NewContext(res http.ResponseWriter, req *http.Request) *Context {
	c := &Context{
		Container: di.NewContainer(),
		Params: make(map[string]string),
		Request: req,
		Response: res,
		Next: func() {},
		NextRoute: func() {},
		Data: make(map[string]interface{}),
	}
	c.Register(c)
	return c
}

// Panic creates a HTTPError and panics with it.
// If the error message is not given, http.StatusText() will be called to generate
// the message based on the status code.
func (c *Context) Panic(status int, message ...string) {
	panic(NewHTTPError(status, message...))
}
