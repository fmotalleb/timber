package server

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/fmotalleb/go-tools/log"

	"github.com/fmotalleb/timber/server/auth"
	"github.com/fmotalleb/timber/server/filesystem"

	"github.com/go-chi/chi/v5"
)

//go:embed static/*
var staticFS embed.FS

func Serve(ctx Context) error {
	l := log.Of(ctx).Named("Serve")
	l.Info("starting server")
	r := chi.NewRouter()
	r.Use(
		withLogger(ctx),
	)
	r.Mount("/static", http.FileServerFS(staticFS))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
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
			w.Write(b)
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

	return http.ListenAndServe(ctx.GetCfg().Listen, r)
}
