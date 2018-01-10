// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentNegotiation(t *testing.T) {
	header := http.Header{}
	header.Set("Accept", "application/json;q=1;v=1")
	req := &http.Request{Header: header}

	offers := []string{"application/json", "application/xml", "application/json;v=1", "application/json;v=2"}
	format := NegotiateContentType(req, offers, "text/html")
	assert.Equal(t, "application/json;v=1", format)
}

func TestContentNegotiation2(t *testing.T) {
	header := http.Header{}
	header.Set("Accept", "application/json;q=0.6;v=1,application/json;v=2")
	req := &http.Request{Header: header}

	offers := []string{"application/json", "application/xml", "application/json;v=1", "application/json;v=2"}
	format := NegotiateContentType(req, offers, "text/html")
	assert.Equal(t, "application/json;v=2", format)
}

func TestContentNegotiation3(t *testing.T) {
	header := http.Header{}
	header.Set("Accept", "*/*,application/xml")
	req := &http.Request{Header: header}

	offers := []string{"application/json", "application/xml", "application/json;v=1", "application/json;v=2"}
	format := NegotiateContentType(req, offers, "text/html")
	assert.Equal(t, "application/xml", format)
}

func TestAccept(t *testing.T) {
	header := http.Header{}
	header.Set("Accept", "application/json;  q=1 ; v=1,")
	req := &http.Request{Header: header}
	mtypes := AcceptMediaTypes(req)

	assert.Equal(t, float64(1), mtypes[0].Weight)
	assert.Equal(t, "application", mtypes[0].Type)
	assert.Equal(t, "json", mtypes[0].Subtype)
	assert.Equal(t, map[string]string{"v": "1", "q": "1"}, mtypes[0].Parameters)
}

func TestAcceptMultiple(t *testing.T) {
	header := http.Header{}
	header.Set("Accept", "application/json;q=1;v=1, application/json;v=2,   text/html")
	req := &http.Request{Header: header}

	mtypes := AcceptMediaTypes(req)

	assert.Equal(t, float64(1), mtypes[0].Weight)
	assert.Equal(t, "application", mtypes[0].Type)
	assert.Equal(t, "json", mtypes[0].Subtype)
	assert.Equal(t, map[string]string{"v": "1", "q": "1"}, mtypes[0].Parameters)

	assert.Equal(t, float64(1), mtypes[1].Weight)
	assert.Equal(t, "application", mtypes[1].Type)
	assert.Equal(t, "json", mtypes[1].Subtype)
	assert.Equal(t, map[string]string{"v": "2"}, mtypes[1].Parameters)

	assert.Equal(t, float64(1), mtypes[2].Weight)
	assert.Equal(t, "text", mtypes[2].Type)
	assert.Equal(t, "html", mtypes[2].Subtype)
	assert.Equal(t, map[string]string{}, mtypes[2].Parameters)
}

func TestAcceptElaborate(t *testing.T) {
	a := `text/plain; q=0.5, text/html, 
          text/x-dvi; q=0.8, text/x-c`

	header := http.Header{}
	header.Set("Accept", a)
	req := &http.Request{Header: header}
	mtypes := AcceptMediaTypes(req)

	assert.Equal(t, float64(0.5), mtypes[0].Weight)
	assert.Equal(t, "text", mtypes[0].Type)
	assert.Equal(t, "plain", mtypes[0].Subtype)

	assert.Equal(t, float64(1), mtypes[1].Weight)
	assert.Equal(t, "text", mtypes[1].Type)
	assert.Equal(t, "html", mtypes[1].Subtype)

	assert.Equal(t, float64(0.8), mtypes[2].Weight)
	assert.Equal(t, "text", mtypes[2].Type)
	assert.Equal(t, "x-dvi", mtypes[2].Subtype)

	assert.Equal(t, float64(1), mtypes[3].Weight)
	assert.Equal(t, "text", mtypes[3].Type)
	assert.Equal(t, "x-c", mtypes[3].Subtype)
}
