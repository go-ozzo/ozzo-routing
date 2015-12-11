# ozzo-routing

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-routing?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-routing)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-routing.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-routing)
[![Coverage](http://gocover.io/_badge/github.com/go-ozzo/ozzo-routing)](http://gocover.io/github.com/go-ozzo/ozzo-routing)

ozzo-routing is a Go package that supports request routing and processing for Web applications.
It has the following features:

* middleware pipeline architecture, similar to that in the [Express framework](http://expressjs.com).
* highly extensible through pluggable handlers (middlewares)
* modular code organization through route grouping
* dependency injection for handler parameters
* URL path parameters
* error handling
* compatible with `http.Handler` and `http.HandlerFunc`

## Requirements

Go 1.2 or above.

## Installation

Run the following command to install the package:

```
go get github.com/go-ozzo/ozzo-routing
```

## Getting Started

Create a `server.go` file with the following content:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
)

func main() {
	r := routing.NewRouter()

	// install commonly used middlewares
	r.Use(
		routing.AccessLogger(log.Printf),
		routing.TrailingSlashRemover(http.StatusMovedPermanently),
	)

	// set up routes and handlers
	r.Get("", func() string {
		return "Go ozzo!"
	})
	r.Get("/users", func() string {
		return "getting users"
	})
	r.Group("/admin", func(gr *routing.Router) {
        gr.Post("/users", func() string {
            return "creating users"
        })
        gr.Delete("/users", func() string {
            return "deleting users"
        })
	})

	// handle requests that don't match any route
	r.Use(routing.NotFoundHandler())

	// handle errors triggered by handlers
	r.Error(routing.ErrorHandler(nil))

	// hook up the router and start up a Go Web server
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
```

Now run the following command to start the Web server:

```
go run server.go
```

You should be able to access URLs such as `http://localhost:8080`, `http://localhost:8080/users`.


## Routing Tree

ozzo-routing works by building a *routing tree* and dispatching HTTP requests to the handlers on this tree.

A leaf node on the routing tree is called a *route*, while a non-leaf node is a *router*. On each node
(either a leaf or a non-leaf node), there is a list of *handlers* (aka middlewares) which contain the custom logic
for handling HTTP requests.

Dispatching an incoming HTTP request starts from the root of the routing tree in a depth-first traversal.
The HTTP method and the URL path are used to match against the encountering nodes. The handlers on the matching
nodes will be invoked according to the node order and handler order. A handler should call either
`Context.Next()` or `Context.NextRoute()` to pass the control to the next eligible handler. Otherwise,
the request handling is considered complete and no further handlers will be invoked.

To build a routing tree, first call `routing.NewRouter()` to create the root node. Then call `Router.To()`, `Router.Get()`,
`Router.Post()`, etc., to create leaf nodes, or call `Router.Group()` to create a non-leaf node. For example,

```go
// root
r := routing.NewRouter()

// leaves (routes)
r.Get("", handler1, handler2, ...)
r.Get("/users", handler1, handler2, ...)

// an internal node (child routers)
r.Group("/admin", func(r *routing.Router) {
    // leaves under the internal node
    r.Post("/users", handler1, handler2, ...)
    r.Delete("/users", handler1, handler2, ...)
})
```

Because `Router` implements `http.Handler`, it can be readily used to serve subtrees on existing Go servers.
For example,

```go
http.Handle("/", r)
http.ListenAndServe(":8080", nil)
```


## Routes

A route has a path pattern that is used to match the URL path of incoming requests. Only requests matching the pattern
may be dispatched by the route. For example, a pattern `/users` matches any request whose URL path is `/users`.
Regular expression can be used in the pattern. For example, a pattern `/users/\\d+` matches URL path `/users/123`,
but not `/users` or `/users/abc`.

Optionally, a route may have one or multiple HTTP methods (e.g. `GET`, `POST`) so that only requests using one of
those HTTP methods may be dispatched by the route.

A route is usually associated with one or multiple handlers. When a route matches and dispatches a request, its handlers
will be called.

You can create and add a route to a routing tree by calling `Router.To()` or one of its shortcut methods, such as `Router.Use()`,
`Router.Get()`. For example,

```go
r := routing.New()

r.To("GET /users", func() { })

// or equivalently using the Get() shortcut
r.Get("/users", func() { })
```

The above code adds a route that matches URL path `/users` and only applies to the GET HTTP method. You may
also call `Post()`, `Put()`, `Patch()`, `Head()`, or `Options()` to deal with other common HTTP methods.

If a route should match multiple HTTP methods, you can use the syntax like shown below:

```go
// only match GET or POST
r.To("GET,POST /users", func() { })

// match any HTTP method
r.To("/users", func() { })
```

If a route should match *any request*, call `Router.Use()` like the following:

```go
r.Use(func() { })
```


### URL Parameters

The path pattern specified for a route can be used to capture URL parameters by embedding tokens in the format
of `<name:pattern>`, where `name` stands for the parameter name, and `pattern` is a regular expression which
the parameter value should match. You can omit the `pattern` part, which means the parameter should match a non-empty
string without any slash character.

When a route matches a URL path, the matching parameters on the URL path will be made available through
`Context.Params`. For example,

```go
r := routing.NewRouter()

r.To("GET /cities/<name>", func (c *routing.Context) {
    fmt.Fprintf(c.Response, "Name: %v", c.Params["name"])
})

r.To("GET /users/<id:\\d+>", func (c *routing.Context) {
    fmt.Fprintf(c.Response, "ID: %v", c.Params["id"])
})
```


## Handlers

Handlers are simply any callable functions. A handler is called when a request is dispatched
to the route or router that the handler is associated with.

Within a handler, you can call `Context.Next()` to pass the control to the next available
handler on the same route or the first handler on the next matching route.
You may also call `Context.NextRoute()` to invoke the first handler of the next matching route.

Usually, handlers serving as filters should call `Context.Next()` so that the next handlers
can get a chance to further process a request. Handlers that are controller actions often do not
call `Context.Next()` because they are the last step of request processing.
`Context.NextRoute()` is often called by handlers to determine if the current route/router
should be used to dispatch the request.

For example,

```go
r := routing.NewRouter()
r.Get("/users", func(c *routing.Context) {
    fmt.Fprintln(c.Response, "/users1 start")
    c.Next()
    fmt.Fprintln(c.Response, "/users1 end")
}, func(c *routing.Context) {
    fmt.Fprintln(c.Response, "/users2 start")
    c.Next()
    fmt.Fprintln(c.Response, "/users2 end")
})

r.Get("/users", func(c *routing.Context) {
    fmt.Fprintln(c.Response, "/users3")
})

r.Get("/users", func(c *routing.Context) {
    fmt.Fprintln(c.Response, "/users4")
})
```

When dispatching the URL path `/users` with the above routing tree, it will output the following text:

```
/users1 start
/users2 start
/users3
/users2 end
/users1 end
```

Note that `/user4` is not displayed because the request dispatching ends after displaying `/user3`.
Also note that the handler outputs are properly nested.


### Context

For each incoming request, a new `routing.Context` instance is created which includes contextual
information for handling the request, such as the current request, response, etc. A handler can get access
to the current `Context` by declaring a `*routing.Context` parameter, like the following:

```go
func (c *routing.Context) {
}
```

Using `Context`, handlers can share data between each other. A simple way is to exploit the `Context.Data` field.
For example, one handler stores `Context.Data["user"]` which can be accessed by another handler. A more advanced
way is to use `Context` as a dependency injection (DI) container. In particular, one handler registers
the data to be shared (e.g. a cache) with `Context` and another handler declares a parameter of the same data type.
Then through automatic dependency injection by `Context`, the latter handler will receive the registered data value
as its parameter. For example,

```go
r := routing.NewRouter()
r.Use(func (c *routing.Context) {
    // use Context.Data to share data
    c.Data["db"] = &Database{}

    // use dependency injection to share data
    c.Register(&Cache{})
})
r.Use(func (c *routing.Context, cache *Cache) {
    // access c.Data["db"]

    // cache is injected
})
```

> Info: When a handler has a `*routing.Context` parameter, its value is also obtained via dependency injection.

### Response and Return Values

Many handlers need to send output in response. This can be done using the following code:

```go
func (c *routing.Context) {
    fmt.Fprint(c.Response, "Hello world")
}
```

An alternative way is to set the output as the return value of a handler. For example, the above code
can be rewritten as follows:

```go
func () string {
    return "Hello world"
}
```

You can return data of arbitrary structure, not just a string. The router will format the return data
into a string by calling `fmt.Fprint()`. You may also customize the data formatting by replacing
`Context.Response` with a response object that implements the `DataWriter` interface.


### Built-in Handlers

ozzo-routing comes with a few commonly used handlers:

* `routing.ErrorHandler`: an error handler
* `routing.NotFoundHandler`: a handler triggering 404 HTTP error
* `routing.TrailingSlashRemover`: a handler removing the trailing slashes from the request URL
* `routing.AccessLogger`: a handler that records an entry for every incoming request
* `routing.StaticFile`: a handler that serves the content of the specified file as the response

These handlers may be used like the following:

```go
r := routing.NewRouter()

r.Use(
    routing.AccessLogger(log.Printf),
    routing.TrailingSlashRemover(http.StatusMovedPermanently),
)

// ... register routes and handlers

r.Use(routing.NotFoundHandler())

r.Error(routing.ErrorHandler(nil))
```

Additional handlers related with RESTful API services may be found in the
[ozzo-rest Go package](https://github.com/go-ozzo/ozzo-rest).


### Third-party Handlers

ozzo-routing supports third-party `http.HandlerFunc` and `http.Handler` handlers. Adapters are provided
to make using third-party handlers an easy task. For example,

```go
r := routing.NewRouter()

// using http.HandlerFunc
r.Use(routing.HTTPHandlerFunc(http.NotFound))

// using http.Handler
r.Use(routing.HTTPHandler(http.NotFoundHandler))
```

## Route Groups

Routes matching the same URL path prefix can be grouped together by calling `Router.Group()`. The support for route
groups enables modular architecture of your application. For example, you can have an `admin` module which uses
the group of the routes having `/admin` as their common URL path prefix. The corresponding routing can be set up
like the following:

```go
r := routing.NewRouter()

// ...other routes...

// the /admin route group
r.Group("/admin", function(gr *routing.Router) {
    gr.Post("/users", func() { })
    gr.Delete("/users", func() { })
    // ...
})
```

Note that when you are creating a route within a route group, the common URL path prefix should be removed
from the path pattern, like shown in the above example.

You can create multiple levels of route groups. In fact, as we have explained earlier, the whole routing system
is a tree structure, which allows you to organize your code in a multilevel modular fashion.


## Error Handling

ozzo-routing supports error handling via error handlers. An error handler is a handler registered
by calling the `Router.Error()` method. When a panic happens in a handler, the router will recover
from it and call the error handlers registered after the current route. Any normal handlers in between
will be skipped.

Error handlers can obtain the error information from `Context.Error`. Like normal handlers,
error handlers also get their parameter values through dependency injection. For example,

```go
r := routing.NewRouter()

// ...register routes and handlers

r.Error(func(c *routing.Context) {
    fmt.Println(c.Error)
})
```

When there are multiple error handlers, `Context.Next()` may be called in one error handler to
pass the control to the next error handler.

For convenience, `Context` provides a method named `Panic()` to simplify the way of triggering an HTTP error.
For example,

```go
func (c *routing.Context) {
    c.Panic(http.StatusNotFound)
    // equivalent to the following code
    // panic(routing.NewHTTPError(http.StatusNotFound))
}
```


## MVC Implementation

ozzo-routing can be used to easily implement the controller part of the MVC pattern.
For example,

```go
// server.go file:
...
r := routing.NewRouter()
...
r.Group("/users", users.Routes)
...

// users/controller.go file:
package users
...
func Routes(r *routing.Router) {
	r.Get("", Controller.index)
	r.Get("/<id:\\d+>", Controller.view)
	...
}

type Controller struct {
	*routing.Context `inject`
}

func (c Controller) index() string {
	return "index"
}

func (c Controller) view() string {
	return "view" + c.Params["id"]
}

...
```

## Credits

ozzo-routing has referenced [Express](http://expressjs.com/), [Martini](https://github.com/go-martini/martini),
and many other similar projects.
