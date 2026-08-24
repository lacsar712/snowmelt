package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/lacsar712/snowmelt/internal/web/api"
)

//go:embed static/*
var staticFS embed.FS

func Handler(server *api.Server) (http.Handler, error) {
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", server.Routes()))
	mux.Handle("/", http.FileServer(http.FS(static)))
	return mux, nil
}
