// Package web embeds the built frontend assets and serves them via an
// http.Handler with SPA-style fallback to index.html.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded frontend.
// Paths that look like file requests (i.e. have a recognizable extension)
// are served from the FS directly. Anything else falls back to index.html
// so client-side routing keeps working after a hard refresh.
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasFileExtension(r.URL.Path) {
			fileServer.ServeHTTP(w, r)
			return
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	}), nil
}

func hasFileExtension(p string) bool {
	idx := strings.LastIndex(p, ".")
	if idx == -1 {
		return false
	}
	return !strings.Contains(p[idx:], "/")
}
