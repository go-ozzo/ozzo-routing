package request

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "context"

    "github.com/ltick/tick-routing"
    "github.com/stretchr/testify/assert"
    "time"
)

func TestTimeout(t *testing.T) {
    h := Timeout(1*time.Second)

    res := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/", nil)
    c := routing.NewContext(res, req, handler1)
    err := h(context.Background(), c)
    assert.Nil(t, err)
    assert.Nil(t, c.Next())
    assert.Equal(t, http.StatusRequestTimeout, res.Code)
    assert.Equal(t, "Request Timeout", res.Body.String())
}

func TestCustomTimeout(t *testing.T) {
    h := Timeout(1*time.Second, func(ctx context.Context, c *routing.Context) error {
        return routing.NewHTTPError(http.StatusRequestTimeout, "Custom Request Timeout")
    })

    res := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/", nil)
    c := routing.NewContext(res, req, handler1)
    err := h(context.Background(), c)
    assert.Nil(t, err)
    assert.Nil(t, c.Next())
    assert.Equal(t, http.StatusRequestTimeout, res.Code)
    assert.Equal(t, "Custom Request Timeout", res.Body.String())
}

func TestCombinedTimeout(t *testing.T) {
    h := Timeout(3*time.Second)

    res := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/", nil)
    c := routing.NewContext(res, req, handler1, handler2)
    err := h(context.Background(), c)
    assert.Nil(t, err)
    assert.Nil(t, c.Next())
    assert.Equal(t, http.StatusOK, res.Code)
    assert.Equal(t, "handler1 Done!Request Timeout", res.Body.String())
}

func TestNoTimeout(t *testing.T) {
    h := Timeout(3*time.Second)

    res := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/", nil)
    c := routing.NewContext(res, req, handler2)
    err := h(context.Background(), c)
    assert.Nil(t, err)
    assert.Nil(t, c.Next())
    assert.Equal(t, http.StatusOK, res.Code)
    assert.Equal(t, "handler2 Done!", res.Body.String())
}

func handler1(ctx context.Context, c *routing.Context) error {
    time.Sleep(2*time.Second)
    c.Write("handler1 Done!")
    return nil
}

func handler2(ctx context.Context, c *routing.Context) error {
    time.Sleep(2*time.Second)
    c.Write("handler2 Done!")
    return nil
}
