package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nuvora-core/nuvora/internal/config"
	"github.com/nuvora-core/nuvora/internal/security"
	"github.com/nuvora-core/nuvora/internal/store"
)

func TestListModels(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := config.Init(); err != nil {
		t.Fatalf("config.Init() error: %v", err)
	}

	if err := store.ConnectDB(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("store.ConnectDB() error: %v", err)
	}

	t.Cleanup(func() {
		if err := store.CloseDB(); err != nil {
			t.Fatalf("store.CloseDB() error: %v", err)
		}
	})

	if err := store.CreateTables(); err != nil {
		t.Fatalf("store.CreateTables() error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/models")
		}

		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong authorization header")
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "z-model"},
				{"id": "a-model"},
				{"id": ""},
			},
		})
	}))

	t.Cleanup(server.Close)

	encryptedKey, err := security.EncryptAPIKey("test-key")

	if err != nil {
		t.Fatalf("security.EncryptAPIKey() error: %v", err)
	}

	_, err = store.DB.Exec(`
		INSERT INTO providers (name, base_url, api_key) VALUES ('groq', ?, ?)
	`, server.URL, encryptedKey)

	if err != nil {
		t.Fatalf("insert provider error: %v", err)
	}

	models, err := ListModels(context.Background(), "")

	if err != nil {
		t.Fatalf("ListModels() error: %v", err)
	}

	want := []string{"a-model", "z-model"}

	if !reflect.DeepEqual(models, want) {
		t.Errorf("ListModels() = %v, want %v", models, want)
	}
}
