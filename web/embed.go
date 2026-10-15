// Package web embeds the built frontend and serves it as a single-page app.
package web

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed all:dist
var dist embed.FS

// Handler serves the embedded frontend. Unknown paths fall back to
// index.html so client-side routing keeps working.
func Handler() (http.Handler, error) {
	assets, err := fs.Sub(dist, "dist")

	if err != nil {
		return nil, fmt.Errorf("web: open dist: %w", err)
	}

	if _, err := fs.Stat(assets, "index.html"); err != nil {
		return nil, errors.New(`web: dist/index.html not found, build the frontend first: "npm run build" in web/`)
	}

	fileServer := http.FileServerFS(assets)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		if file == "" || !fileExists(assets, file) {
			// The entry point must always be revalidated so fresh
			// builds are picked up.
			w.Header().Set("Cache-Control", "no-cache")

			serveIndex(w, r, assets)

			return
		}

		fileServer.ServeHTTP(w, r)
	}), nil
}

func serveIndex(w http.ResponseWriter, r *http.Request, assets fs.FS) {
	index, err := fs.ReadFile(assets, "index.html")

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(index))
}

func fileExists(assets fs.FS, name string) bool {
	info, err := fs.Stat(assets, name)

	return err == nil && !info.IsDir()
}
