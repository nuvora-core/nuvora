package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/nuvora-core/nuvora/internal/ai"
)

// ListModels lists the models offered by the connected provider.
// The optional provider query parameter picks a specific provider.
func ListModels(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	providerName := r.URL.Query().Get("provider")

	models, err := ai.ListModels(r.Context(), providerName)

	if errors.Is(err, ai.ErrNoProvider) {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "no provider connected",
		})

		return
	}

	if err != nil {
		fmt.Println("providers: list models error:", err)

		w.WriteHeader(http.StatusBadGateway)

		json.NewEncoder(w).Encode(map[string]any{
			"error":  "failed to load models",
			"reason": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"models": models,
	})
}
