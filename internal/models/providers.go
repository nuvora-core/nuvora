package models

type ConnectProviderReq struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

// Provider is the safe representation of a connected provider.
// The api key is intentionally left out.
type Provider struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}
