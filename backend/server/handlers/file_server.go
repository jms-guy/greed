package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

//Sets up a http.FileServer handler for serving static files from a http.FileSystem
func (app *AppServer) FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.Contains(path, "{}*") {
		app.Logger.Log(
			"level", "error",
			"msg", "FileServer does not permit any URL parameters",
		)
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}