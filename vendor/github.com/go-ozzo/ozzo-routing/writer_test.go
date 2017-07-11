package routing

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDataWriter(t *testing.T) {
	res := httptest.NewRecorder()
	err := DefaultDataWriter.Write(res, "abc")
	assert.Nil(t, err)
	assert.Equal(t, "abc", res.Body.String())

	res = httptest.NewRecorder()
	err = DefaultDataWriter.Write(res, []byte("abc"))
	assert.Nil(t, err)
	assert.Equal(t, "abc", res.Body.String())

	res = httptest.NewRecorder()
	err = DefaultDataWriter.Write(res, 123)
	assert.Nil(t, err)
	assert.Equal(t, "123", res.Body.String())

	res = httptest.NewRecorder()
	err = DefaultDataWriter.Write(res, nil)
	assert.Nil(t, err)
	assert.Equal(t, "", res.Body.String())

	res = httptest.NewRecorder()
	c := &Context{}
	c.init(res, nil)
	assert.Nil(t, c.Write("abc"))
	assert.Equal(t, "abc", res.Body.String())
}
