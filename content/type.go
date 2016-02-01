// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package content provides content negotiation handlers for the ozzo routing package.
package content

import (
	"encoding/json"
	"encoding/xml"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/golang/gddo/httputil"
	"net/http"
)

const (
	JSON = "application/json"
	XML  = "application/xml"
	HTML = "text/html"
)

// Formatter is a function setting response content type and returning a routing.SerializeFunc for writing data.
type Formatter func(http.ResponseWriter) routing.SerializeFunc

// Formatters lists all supported content types and the corresponding formatters.
// By default, JSON, XML, and HTML are supported. You may modify this variable before calling TypeNegotiator
// to customize supported formatters.
var Formatters = map[string]Formatter{
	JSON: JSONFormatter,
	XML:  XMLFormatter,
	HTML: HTMLFormatter,
}

// TypeNegotiator returns a content type negotiation handler.
//
// The method takes a list of response MIME types that are supported by the application.
// The negotiator will determine the best response MIME type to use by checking the Accept request header.
// If no match is found, the first MIME type will be used.
//
// The negotiator will set the "Content-Type" response header as the chosen MIME type. It will also set
// routing.Context.Write to be the function that would serialize the given data in the appropriate format.
//
// If you do not specify any supported MIME types, the negotiator will use "text/html" as the response MIME type.
func TypeNegotiator(formats ...string) routing.Handler {
	if len(formats) == 0 {
		formats = []string{HTML}
	}
	for _, format := range formats {
		if _, ok := Formatters[format]; !ok {
			panic(format + " is not supported")
		}
	}
	defaultFormat := formats[0]

	return func(c *routing.Context) error {
		format := httputil.NegotiateContentType(c.Request, formats, defaultFormat)
		c.Serialize = Formatters[format](c.Response)
		return nil
	}
}

// JSONFormatter sets the "Content-Type" response header as "application/json" and returns a routing.WriteFunc that writes the given data in JSON format.
func JSONFormatter(res http.ResponseWriter) routing.SerializeFunc {
	res.Header().Set("Content-Type", "application/json")
	return json.Marshal
}

// XMLFormatter sets the "Content-Type" response header as "application/xml; charset=UTF-8" and returns a routing.WriteFunc that writes the given data in XML format.
func XMLFormatter(res http.ResponseWriter) routing.SerializeFunc {
	res.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	return xml.Marshal
}

// HTMLFormatter sets the "Content-Type" response header as "text/html; charset=UTF-8" and returns a routing.WriteFunc that writes the given data in a byte stream.
func HTMLFormatter(res http.ResponseWriter) routing.SerializeFunc {
	res.Header().Set("Content-Type", "text/html; charset=UTF-8")
	return routing.Serialize
}
