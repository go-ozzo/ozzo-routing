package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FA struct {
	A1 string
	A2 int
}

type FB struct {
	B1 string
	B2 bool
	B3 float64
}

func TestReadForm(t *testing.T) {
	var a struct {
		X1 string `form:"x1"`
		FA
		X2 int
		B  *FB
		FB `form:"c"`
		c  int
	}
	values := map[string][]string{
		"x1":   []string{"abc", "123"},
		"A1":   []string{"a1"},
		"x2":   []string{"1", "2"},
		"B.B1": []string{"b1", "b2"},
	}
	err := ReadForm(values, &a)
	assert.Nil(t, err)
	assert.Equal(t, "abc", a.X1)
	assert.Equal(t, "a1", a.A1)
	assert.Equal(t, 0, a.X2)
}
