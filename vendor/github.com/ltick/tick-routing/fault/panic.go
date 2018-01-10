package fault

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/go-ozzo/ozzo-routing"
)

// PanicHandler returns a handler that recovers from panics happened in the handlers following this one.
// When a panic is recovered, it will be converted into an error and returned to the parent handlers.
//
// A log function can be provided to log the panic call stack information. If the log function is nil,
// no message will be logged.
//
//     import (
//         "log"
//         "github.com/go-ozzo/ozzo-routing"
//         "github.com/go-ozzo/ozzo-routing/fault"
//     )
//
//     r := routing.New()
//     r.Use(fault.ErrorHandler(log.Printf))
//     r.Use(fault.PanicHandler(log.Printf))
func PanicHandler(logf LogFunc) routing.Handler {
	return func(c *routing.Context) (err error) {
		defer func() {
			if e := recover(); e != nil {
				if logf != nil {
					logf("recovered from panic:%v", getCallStack(4))
				}
				var ok bool
				if err, ok = e.(error); !ok {
					err = fmt.Errorf("%v", e)
				}
			}
		}()

		return c.Next()
	}
}

// getCallStack returns the current call stack information as a string.
// The skip parameter specifies how many top frames should be skipped.
func getCallStack(skip int) string {
	buf := new(bytes.Buffer)
	for i := skip; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "\n%s:%d", file, line)
	}
	return buf.String()
}
