package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/auth"
	"github.com/fmotalleb/timber/server/filesystem"

	"github.com/go-chi/chi/v5"
)

//go:embed static/*
var staticFS embed.FS

const readHeaderTimeout = 3 * time.Second

// Serve starts the HTTP server.
func Serve(ctx Context) error {
	l := log.Of(ctx).Named("Serve")
	l.Info("starting server")
	r := chi.NewRouter()
	r.Use(
		withLogger(ctx),
	)

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })
	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(
			auth.WithBasicAuth(ctx.GetCfg()),
			auth.PermissionCheck,
		)
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok {
				return
			}
			b, _ := json.Marshal(user)
			if _, err := w.Write(b); err != nil {
				log.Of(r.Context()).Error("failed to write response", zap.Error(err))
			}
		})
		r.Get(
			"/filesystem/ls",
			filesystem.Ls,
		)
		r.Get(
			"/filesystem/cat",
			filesystem.Cat,
		)
		r.Get(
			"/filesystem/head",
			filesystem.Head,
		)
		r.Get(
			"/filesystem/tail",
			filesystem.Tail,
		)
	})
	rootFs, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	r.Mount("/", http.FileServerFS(rootFs))
	server := &http.Server{
		Addr:              ctx.GetCfg().Listen,
		ReadHeaderTimeout: readHeaderTimeout,
		Handler:           r,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return server.Shutdown(ctx)
	}
}
