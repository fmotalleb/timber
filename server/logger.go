package server

import (
	"net/http"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func withLogger(ctx Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			l := log.Of(r.Context()).
				Named("Server").
				Named("Request").
				With(
					zap.Time("started", t1),
					zap.String("client", r.RemoteAddr),
					zap.String("uri", r.RequestURI),
				)
			defer func() {
				status := ww.Status()
				if status >= 400 {
					l.Warn(
						"request finished with error",
						zap.Int("status", status),
						zap.Int("response_length", ww.BytesWritten()),
						zap.Any("headers", ww.Header()),
						zap.Duration("time", time.Since(t1)),
					)
				} else {
					l.Debug(
						"request finished",
						zap.Int("status", status),
						zap.Int("response_length", ww.BytesWritten()),
						zap.Any("headers", ww.Header()),
						zap.Duration("time", time.Since(t1)),
					)
				}
			}()
			newCtx := log.WithLogger(r.Context(), l)
			next.ServeHTTP(ww, r.WithContext(newCtx))
		}
		return http.HandlerFunc(fn)
	}
}
