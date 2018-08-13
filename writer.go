package routing

import (
	"net/http"
    "fmt"
)

// DataWriter is used by Context.Write() to write arbitrary data into an HTTP response.
type DataWriter interface {
	// SetHeader sets necessary response headers.
	SetHeader(http.ResponseWriter)
	// Write writes the given data into the response.
	Write(http.ResponseWriter, interface{}) (int, error)
}

// DefaultDataWriter writes the given data in an HTTP response.
// If the data is neither string nor byte array, it will use fmt.Fprint() to write it into the response.
var DefaultDataWriter DataWriter = &dataWriter{}

type dataWriter struct{}

func (w *dataWriter) SetHeader(res http.ResponseWriter) {}

func (w *dataWriter) Write(res http.ResponseWriter, data interface{}) (n int, err error) {
	switch data.(type) {
	case []byte:
		dataByte := data.([]byte)
		n, err = res.Write(dataByte)
	case string:
		dataByte := []byte(data.(string))
		n, err = res.Write(dataByte)
	default:
		if data != nil {
			n, err = fmt.Fprint(res, data)
		}
	}
	return n, err
}
