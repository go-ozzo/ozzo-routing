package routing

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterNotFound(t *testing.T) {
	r := New()
	h := func(c *Context) error {
		fmt.Fprint(c.Response, "ok")
		return nil
	}
	r.Get("/users", h)
	r.Post("/users", h)
	r.NotFound(MethodNotAllowedHandler, NotFoundHandler)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET, OPTIONS, POST", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, http.StatusMethodNotAllowed, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET, OPTIONS, POST", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, http.StatusOK, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/posts", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, http.StatusNotFound, res.Code, "HTTP status code")
}

func TestRouterUse(t *testing.T) {
	r := New()
	assert.Equal(t, 2, len(r.notFoundHandlers))
	r.Use(NotFoundHandler)
	assert.Equal(t, 3, len(r.notFoundHandlers))
}

func TestRouterRoute(t *testing.T) {
	r := New()
	r.Get("/users").Name("users")
	assert.NotNil(t, r.Route("users"))
	assert.Nil(t, r.Route("users2"))
}

func TestRouterAdd(t *testing.T) {
	r := New()
	assert.Equal(t, 0, r.maxParams)
	r.add("GET", "/users/<id>", nil)
	assert.Equal(t, 1, r.maxParams)
}

func TestRouterFind(t *testing.T) {
	r := New()
	r.add("GET", "/users/<id>", []Handler{NotFoundHandler})
	pvalues := make([]string, 10)
	handlers, pnames := r.find("GET", "/users/1", pvalues)
	assert.Equal(t, 1, len(handlers))
	if assert.Equal(t, 1, len(pnames)) {
		assert.Equal(t, "id", pnames[0])
	}
	assert.Equal(t, "1", pvalues[0])
}

func TestRouterHandleError(t *testing.T) {
	r := New()
	res := httptest.NewRecorder()
	c := &Context{Response: res}
	r.handleError(c, errors.New("abc"))
	assert.Equal(t, http.StatusInternalServerError, res.Code)

	res = httptest.NewRecorder()
	c = &Context{Response: res}
	r.handleError(c, NewHTTPError(http.StatusNotFound))
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestHTTPHandler(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := NewContext(res, req)

	h1 := HTTPHandlerFunc(http.NotFound)
	assert.Nil(t, h1(c))
	assert.Equal(t, http.StatusNotFound, res.Code)

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = NewContext(res, req)
	h2 := HTTPHandler(http.NotFoundHandler())
	assert.Nil(t, h2(c))
	assert.Equal(t, http.StatusNotFound, res.Code)
}
