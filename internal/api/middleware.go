package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

type traceKey struct{}

var traceSequence atomic.Uint64

func requestTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("trace-%06d", traceSequence.Add(1))
		w.Header().Set("X-Trace-ID", id)
		ctx := context.WithValue(r.Context(), traceKey{}, id)
		started := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		log.Printf("request trace=%s method=%s path=%s duration=%s", id, r.Method, r.URL.Path, time.Since(started).Round(time.Millisecond))
	})
}
func detachFrameCancellation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/frames") {
			r = r.WithContext(context.WithoutCancel(r.Context()))
		}
		next.ServeHTTP(w, r)
	})
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				log.Printf("panic trace=%s value=%v stack=%s", traceFrom(r.Context()), value, debug.Stack())
				writeError(w, http.StatusInternalServerError, "internal server failure")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func traceFrom(ctx context.Context) string { value, _ := ctx.Value(traceKey{}).(string); return value }
