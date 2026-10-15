package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/nuvora-core/nuvora/internal/models"
	"github.com/nuvora-core/nuvora/internal/store"
	"modernc.org/sqlite"
)

func CreateNewProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req models.NewProjectReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid request",
		})

		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name is required",
		})
		return
	}

	if len(req.Name) > 75 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name is too long",
		})
		return
	}

	if store.DB == nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database not connected",
		})

		return
	}

	info, err := os.Stat(req.RootPath)

	if err == nil && info.IsDir() {
		_, err = store.DB.Exec(
			`
			INSERT INTO projects (name, root_path)
			VALUES (?,?);
			`,
			req.Name,
			req.RootPath,
		)

		if err != nil {
			var sqliteErr *sqlite.Error

			if errors.As(err, &sqliteErr) {
				if sqliteErr.Code() == 2067 {
					w.WriteHeader(http.StatusConflict)

					json.NewEncoder(w).Encode(map[string]any{
						"error": "There is already a project with this name.",
					})

					return
				}
			}

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "database error",
			})

			return
		}

		w.WriteHeader(201)

		json.NewEncoder(w).Encode(map[string]any{
			"message": "project created",
		})

		return
	} else {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})

		return
	}
}

func MyProjects(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database not connected",
		})

		return
	}

	rows, err := store.DB.QueryContext(r.Context(), `
		SELECT id, name, root_path, created_at FROM projects ORDER BY id
	`)

	if err != nil {
		fmt.Println("projects: list projects error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load projects",
		})

		return
	}

	defer rows.Close()

	projects := []models.Project{}

	for rows.Next() {
		var project models.Project

		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.RootPath,
			&project.CreatedAt,
		); err != nil {
			fmt.Println("projects: scan project error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "failed to load projects",
			})

			return
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("projects: iterate projects error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load projects",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"projects": projects,
	})
}
