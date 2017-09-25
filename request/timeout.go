// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fault provides a panic and error handler for the ozzo routing package.
package request

import (
    "github.com/ltick/tick-routing"
    "context"
    "time"
)

func Timeout(timeoutDuration time.Duration, handlers ...routing.Handler) routing.Handler {
    return func(ctx context.Context, c *routing.Context) error {
        c.TimeoutDuration = timeoutDuration
        if len(handlers) > 0 {
            c.TimeoutHandlers = handlers
        }
        if c.TimeoutDuration != 0*time.Second {
            c.Ctx, c.CancelFunc = context.WithTimeout(c.Ctx, c.TimeoutDuration)
        }
        return nil
    }
}
