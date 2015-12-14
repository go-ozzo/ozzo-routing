# ozzo-routing

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-routing?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-routing)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-routing.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-routing)
[![Coverage](http://gocover.io/_badge/github.com/go-ozzo/ozzo-routing)](http://gocover.io/github.com/go-ozzo/ozzo-routing)

ozzo-routing это Go пакет, обспечивающий поддержку роутинга и обработки для Web приложений.
Он включает в себя:

* middleware pipeline architecture, подобную [Express framework](http://expressjs.com).
* высокую расширяемость при помощи подключемых обработчиков (middlewares)
* модульную организацию кода через группировку маршрутов
* внедрение зависимостей (dependency injection) для параметров handlerов
* воможность использовать URL с параметрами
* файловый сервер для статического контента (static file server): js, png, jpg и так далее.
* обработчик ошибок error handling
* совместимость с нативными `http.Handler` и `http.HandlerFunc`

## Требования

Go 1.2 или выше.

## Установка

Выполните указанные команды для установки пакета:

```
go get github.com/go-ozzo/ozzo-routing
```

## С чего начать

Создайте файл `server.go` с таким содержимым:

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

Теперь выполите команды для запуска Web сервера:

```
go run server.go
```

Теперь вы можете получить доступ по адресу URLs таким как `http://localhost:8080`, `http://localhost:8080/users`.


## Дерево маршрутизции

ozzo-routing работает при помощи создания дерева маршрутов *routing tree* распределяя HTTP запросы по обработчикам на этом дереве.

Конечный узел на дереве маршрутов называется *route*, в то время как промежуточныый называется *router*. На каждом узле
(конечном или промежуточном), имеется список обработчиков *handlers* (также известные как middlewares) которые сожержат логику
для обработки HTTP запросов.

Диспетчеризация входящих HTTP начинается с корня дерева маршрутизации и далее обход идет в глубину.
Методы HTTP и части пути URL используются для соответствия встречным узлам. Обработчики для совпадающих узлов будут вызываться
в соответствии с порядком узла и порядком следования обработчика для этого узла.
Обработчик должен вызывать либо `Context.Next()` или `Context.NextRoute()` чтобы передать управление на следующий разрешенный обаботчик. В противном случае,
обработка запроса считается завершенной и дальнейшие обработчики вызываться не будут.

Для построения дерева маршрутизации, сначала выполните `routing.NewRouter()` для создания корневого узла. Затем выполните `Router.To()`, `Router.Get()`,
`Router.Post()`, и так далее для создания конечных узлов, или вызовете `Router.Group()` для создания промежуточного узла. Например,

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

В следствие того что `Router` имплементирует `http.Handler`, он может быть легко использован для обслуживания поддеревьев на существующих Go серверах.
Например,

```go
http.Handle("/", r)
http.ListenAndServe(":8080", nil)
```


## Маршруты

Маршрут содержит шаблон пути, который используется для соответствия URL и пути входящего запроса. Только запросы соответствующие определенному шаблону могут быть направлены по соответствующему маршруту. Например, шаблон `/users` соответствует любому запросу с URL чей путь соответствует `/users`.
Регулярные выражения также могут быть использованы в шаблоне. Например, шаблон `/users/\\d+` соответствует URL запросу `/users/123`,
но не `/users` или `/users/abc`.

Опчионально маршрут может содержать один или несколько HTTP методов (таких как `GET`, `POST`) таком образом, что только запросы использующие 
один из таких методов HTTP могут быть оправлены по маршруту.

Маршрут обычно ассоциируется с одним или несколькими обработчиками. Когда при разборе запроса обнаружется сопадение маршрутов, 
их обработчики будут вызваны.

Вы можете создать или добавить новый маршрут к дереву маршрутизации с помощью вызова `Router.To()` или одного из его укороченных методов, 
таких как `Router.Use()`, `Router.Get()`. Например,

```go
r := routing.New()

r.To("GET /users", func() { })

// что эквивалентно использованию метода Get()
r.Get("/users", func() { })
```

Приведенный выше код добавляет маршрут, который соответствует пути URL `/users` и применяется только к HTTP методу GET. Также вы можете вызывать
`Post()`, `Put()`, `Patch()`, `Head()`, или `Options()` для вызова прочих HTTP методов.

Если маршрут должен использовать несколько методов HTTP, вы можете использовать синтаксис, как показано ниже:

```go
// соответствует только GET или POST
r.To("GET,POST /users", func() { })

// применяется для обработки любого метода HTTP
r.To("/users", func() { })
```

Если маршрут должен соответствовать *любому запросу*, вызывайте `Router.Use()` как указано ниже:

```go
r.Use(func() { })
```


### параметры URL

Шаблон, определенный для маршрута, может быть использован для сбора параметров из строки URL путем встраивания маркеров в формат в виде `<имя:шаблон>`, 
где `имя` означает имя параметра, и `шаблон` является регулярным выражением, с которым значение параметра должно совпадать. ВЫ можете опустить часть `шаблон`,
что будет означать что параметру должгна соответствовать непустая строка без символа слэша.

Когда маршрут соответствует пути URL, параметры URL будут доступны через `Context.Params`. Например,

```go
r := routing.NewRouter()

r.To("GET /cities/<name>", func (c *routing.Context) {
    fmt.Fprintf(c.Response, "Name: %v", c.Params["name"])
})

r.To("GET /users/<id:\\d+>", func (c *routing.Context) {
    fmt.Fprintf(c.Response, "ID: %v", c.Params["id"])
})
```


## Обработчики

Обработчики - это обычные вызываемые функции. Обработчик вызывается, когда запрос отправляется
на маршрут или маршрутизатор, с которым связан данный обработчик.

В обработчике вы можете вызвать `Context.Next()` чтобы передать управление на следующий доступный обработчик
по тому же самому маршруту или первому обработчику на следующем совпадающем маршруте.
Вы также можете вызвать `Context.NextRoute()` чтобы сразу вызвать обработчик следующего совпадающего маршрута.

Обычно обработчики выступающие в качестве фильтра следует вызывать `Context.Next()` чтобы следующий за ним обработчик смог
обработать запрос. Обработчики, которые являются контроллерами действий обычно не должны вызывать `Context.Next()` т.к. 
они являются последним этапом обработки запроса.
`Context.NextRoute()` часто используются обаботчиками для определеия если ткущий маршрут или роутер 
используется для отправки запроса.

Например,

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

Заметьте, что `/user4` не отображается потому что разбор запроса прекращается после отображения `/user3`.
Также обратите внимание, что выходы обработчик вложены правильно.


### Контекст

Для каждого входящего запроса, создается новый экземпляр `routing.Context` в который включается контекстаная 
информация для входящего запроса, такая как текущий запрос, ответ, и т.д. Обработчик может получить доступ к текущему `Контексту` объявив параметр `*routing.Context`, как описано ниже:

```go
func (c *routing.Context) {
}
```

Используя `Контекст`, обработчики могут обмениваться данными между собой. Простой способ заключается в использовани поля `Context.Data`.
Например один обработчик сохраняет данные в поле `Context.Data["user"]`, которое будет доступно для другого обработчика. Более продвинутый
способ заключается в использовании `Контекста` как инъекции зависимостей с использованием (DI) контейнера. В частности, один обработчик регистрирует
данные которые должны быть доступны для совместного использования (такие как кеш) с `Контекстом` и другой обработчик объявляет параметр тогоже типа данных.
Зачем через внедрение зависимостей `Context`, следующий обработчик сможет принять данные как свой параметр. Например,

```go
r := routing.NewRouter()
r.Use(func (c *routing.Context) {
    // используя Context.Data для обмена данными
    c.Data["db"] = &Database{}

    // используя dependency injection для обмена данными
    c.Register(&Cache{})
})
r.Use(func (c *routing.Context, cache *Cache) {
    // доступ из c.Data["db"]

    // cache уже внедрено и доступно
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
* `routing.Static`: a handler that serves the files under the specified folder as response content
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

## Serving Static Files

Static files can be served through the `routing.Static` or `routing.StaticFile` handler. The former serves files
under the specified directory according to the current request, while the latter serves a single file. For example,

```go
r := routing.NewRouter()
// serves the files under working-dir/web/assets
r.To("/assets(/.*)?", routing.Static("web"))
```


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
