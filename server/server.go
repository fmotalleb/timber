package server

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/fmotalleb/go-tools/log"

	"github.com/fmotalleb/timber/server/auth"

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
		auth.WithBasicAuth(ctx.GetCfg()),
	)
	r.Mount("/static", http.FileServerFS(staticFS))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			return
		}
		b, _ := json.Marshal(user)
		w.Write(b)
	})

	return http.ListenAndServe(ctx.GetCfg().Listen, r)
}
