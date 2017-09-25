package request

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "context"

    "github.com/ltick/tick-routing"
    "github.com/stretchr/testify/assert"
)

func TestCancelHandler(t *testing.T) {
    h := CancelHandler(handler3)

    res := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/", nil)
    c := routing.NewContext(res, req, handler3)
    err := h(context.Background(), c)
    assert.Nil(t, err)
    assert.Nil(t, c.Next())
    assert.Equal(t, http.StatusOK, res.Code)
    assert.Equal(t, "123", res.Body.String())
}

func handler3(ctx context.Context, c *routing.Context) error {
    c.Write("123")
    return nil
}

