package main

import (
	"context"
	"time"
)

func signalContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	parent := context.Background()
	return context.WithTimeout(parent, timeout)
}
