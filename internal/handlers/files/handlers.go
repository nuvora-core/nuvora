package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nuvora-core/nuvora/internal/models"
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

// SaveFile writes new content to an existing file.
func SaveFile(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req models.SaveFileReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid request",
		})

		return
	}

	if req.Path == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "path is required",
		})

		return
	}

	abs, err := filepath.Abs(req.Path)

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

	if len(req.Content) > maxFileSize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "file is too large to save",
		})

		return
	}

	if err := os.WriteFile(abs, []byte(req.Content), info.Mode().Perm()); err != nil {
		fmt.Println("files: save file error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to save file",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"status": "saved",
	})
}
