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
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "ok", res.Body.String(), "response body")
	assert.Equal(t, http.StatusOK, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET, OPTIONS, POST", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, http.StatusMethodNotAllowed, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET, OPTIONS, POST", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, http.StatusOK, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, http.StatusNotFound, res.Code, "HTTP status code")

	r.IgnoreTrailingSlash = true
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "ok", res.Body.String(), "response body")
	assert.Equal(t, http.StatusOK, res.Code, "HTTP status code")
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
	handlers, params := r.Find("GET", "/users/1")
	assert.Equal(t, 1, len(handlers))
	if assert.Equal(t, 1, len(params)) {
		assert.Equal(t, "1", params["id"])
	}
}

func TestRouterNormalizeRequestPath(t *testing.T) {
	tests := []struct{
		path string
		expected string
	}{
		{"/", "/"},
		{"/users", "/users"},
		{"/users/", "/users"},
		{"/users//", "/users"},
		{"///", "/"},
	}
	r := New()
	r.IgnoreTrailingSlash = true
	for _, test := range tests {
		result := r.normalizeRequestPath(test.path)
		assert.Equal(t, test.expected, result)
	}
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
