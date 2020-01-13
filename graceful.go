// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown shuts down the given HTTP server gracefully when receiving an os.Interrupt or syscall.SIGTERM signal.
// It will wait for the specified timeout to stop hanging HTTP handlers.
func GracefulShutdown(hs *http.Server, timeout time.Duration, logFunc func(format string, args ...interface{})) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logFunc("shutting down server with %s timeout", timeout)

	if err := hs.Shutdown(ctx); err != nil {
		logFunc("error while shutting down server: %v", err)
	} else {
		logFunc("server was shut down gracefully")
	}
}
