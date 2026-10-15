package providers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nuvora-core/nuvora/internal/config"
	"github.com/nuvora-core/nuvora/internal/models"
	"github.com/nuvora-core/nuvora/internal/security"
	"github.com/nuvora-core/nuvora/internal/store"
)

func ConnectProvider(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req models.ConnectProviderReq

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

	// Only the supported providers are accepted and each one has a
	// fixed base URL from the config. This way a client cannot point a
	// known provider at an arbitrary host.
	switch req.Name {
	case "groq":
		req.BaseURL = config.GroqBaseURL
	case "openrouter":
		req.BaseURL = config.OpenrouterBaseURL
	}

	if req.BaseURL == "" {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "provider base url is not configured",
		})

		return
	}

	if req.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "api key is required",
		})

		return
	}

	// The API key is never written to the database as plain text.
	// security.EncryptAPIKey returns base64(nonce || ciphertext).
	encryptedKey, err := security.EncryptAPIKey(req.APIKey)
	if err != nil {
		fmt.Println("providers: encrypt api key error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to secure api key",
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

	// base_url is UNIQUE, so connecting the same provider again
	// refreshes the stored (encrypted) key instead of failing.
	_, err = store.DB.ExecContext(r.Context(), `
		INSERT INTO providers (name, base_url, api_key)
		VALUES (?, ?, ?)
		ON CONFLICT(base_url) DO UPDATE SET
			name = excluded.name,
			api_key = excluded.api_key
	`, req.Name, req.BaseURL, encryptedKey)
	if err != nil {
		fmt.Println("providers: save provider error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to save provider",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"status":   "connected",
		"provider": req.Name,
		"base_url": req.BaseURL,
	})
}

func GetYourProviders(
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
		SELECT id,name, base_url FROM providers ORDER BY id
	`)
	if err != nil {
		fmt.Println("providers: list providers error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load providers",
		})

		return
	}

	defer rows.Close()

	providers := []models.Provider{}

	for rows.Next() {
		var provider models.Provider

		if err := rows.Scan(&provider.Name, &provider.BaseURL); err != nil {
			fmt.Println("providers: scan provider error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "failed to load providers",
			})

			return
		}

		// The API key stays encrypted in the database and is never
		// sent back to the client.
		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("providers: iterate providers error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load providers",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"providers": providers,
	})
}
