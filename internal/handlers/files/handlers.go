package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// maxFileSize caps the size of files the viewer can open.
const maxFileSize = 1 << 20 // 1 MiB

// ReadFile returns the text content of a file so it can be shown in
// the project file viewer.
func ReadFile(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "path is required",
		})

		return
	}

	abs, err := filepath.Abs(path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid path",
		})

		return
	}

	info, err := os.Stat(abs)

	if err != nil || info.IsDir() {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "not a file",
		})

		return
	}

	if info.Size() > maxFileSize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "file is too large to open",
		})

		return
	}

	data, err := os.ReadFile(abs)

	if err != nil {
		fmt.Println("files: read file error:", err)

		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "cannot read file",
		})

		return
	}

	if bytes.IndexByte(data, 0) != -1 {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "binary files are not supported",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"path":    abs,
		"name":    filepath.Base(abs),
		"size":    info.Size(),
		"content": string(data),
	})
}
