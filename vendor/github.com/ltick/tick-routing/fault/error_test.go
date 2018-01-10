package fault

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	h := ErrorHandler(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test", res.Body.String())
	assert.Equal(t, "", buf.String())

	buf.Reset()
	h = ErrorHandler(getLogger(&buf), convertError)
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "123", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	h = ErrorHandler(nil)
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())
	assert.Equal(t, "", buf.String())
}

func Test_writeError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req)
	writeError(c, errors.New("abc"))
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req)
	writeError(c, routing.NewHTTPError(http.StatusNotFound, "xyz"))
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "xyz", res.Body.String())
}

func convertError(c *routing.Context, err error) error {
	return errors.New("123")
}
