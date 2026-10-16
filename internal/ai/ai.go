// Package ai talks to the configured provider using the
// OpenAI-compatible chat completions API (Groq, OpenRouter).
package ai

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/nuvora-core/nuvora/internal/security"
	"github.com/nuvora-core/nuvora/internal/store"
	"github.com/nuvora-core/nuvora/internal/tools"
)

// ErrNoProvider is returned when no provider is connected yet.
var ErrNoProvider = errors.New("no provider connected")

// Message is a single chat message sent to the provider.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolCall is a tool call requested by the model.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// apiMessage is a message in the OpenAI chat format, including tool
// calls and tool results.
type apiMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolEvent records one executed tool call for persistence and display.
type ToolEvent struct {
	Display string
	Result  string
}

// ChatResult is the outcome of a chat exchange.
type ChatResult struct {
	Reply      string
	ToolEvents []ToolEvent
}

// defaultModels maps each provider to the model used when the request
// does not specify one.
var defaultModels = map[string]string{
	"groq":       "llama-3.3-70b-versatile",
	"openrouter": "openrouter/auto",
}

const (
	maxToolSteps       = 8
	maxToolResultBytes = 4000
)

type providerInfo struct {
	name    string
	baseURL string
	apiKey  string
}

var client = &http.Client{Timeout: 120 * time.Second}

// ChatWithTools sends the message history to the provider and runs a
// tool calling loop: whenever the model requests a tool, it is
// executed inside projectRoot and the result is fed back, until the
// model produces a final reply. providerName may be empty, in which
// case the first connected provider is used.
func ChatWithTools(
	ctx context.Context,
	providerName string,
	model string,
	history []Message,
	projectRoot string,
) (ChatResult, error) {
	if store.DB == nil {
		return ChatResult{}, errors.New("database not connected")
	}

	provider, err := connectProvider(ctx, providerName)

	if err != nil {
		return ChatResult{}, err
	}

	if model == "" {
		model = defaultModels[provider.name]
	}

	if model == "" {
		return ChatResult{}, fmt.Errorf("no default model for provider %q", provider.name)
	}

	messages := make([]apiMessage, 0, len(history)+maxToolSteps*2)

	for _, message := range history {
		messages = append(messages, apiMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	schemas := tools.Schemas()
	events := []ToolEvent{}

	for step := 0; step < maxToolSteps; step++ {
		reply, err := chatOnce(ctx, provider, model, messages, schemas)

		if err != nil {
			return ChatResult{}, err
		}

		if len(reply.ToolCalls) == 0 {
			return ChatResult{
				Reply:      reply.Content,
				ToolEvents: events,
			}, nil
		}

		messages = append(messages, apiMessage{
			Role:      "assistant",
			Content:   reply.Content,
			ToolCalls: reply.ToolCalls,
		})

		for _, call := range reply.ToolCalls {
			display, result := executeToolCall(ctx, projectRoot, call)

			events = append(events, ToolEvent{
				Display: display,
				Result:  result,
			})

			messages = append(messages, apiMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    result,
			})
		}
	}

	return ChatResult{}, errors.New("ai exceeded the tool call limit")
}

// ListModels returns the sorted model IDs the provider offers.
func ListModels(
	ctx context.Context,
	providerName string,
) ([]string, error) {
	if store.DB == nil {
		return nil, errors.New("database not connected")
	}

	provider, err := connectProvider(ctx, providerName)

	if err != nil {
		return nil, err
	}

	url := strings.TrimSuffix(provider.baseURL, "/") + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("build models request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+provider.apiKey)

	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("models request: %w", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))

	if err != nil {
		return nil, fmt.Errorf("read models response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"models request failed (%d): %s",
			res.StatusCode,
			truncate(strings.TrimSpace(string(body)), 300),
		)
	}

	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode models response: %w", err)
	}

	models := make([]string, 0, len(parsed.Data))

	for _, model := range parsed.Data {
		if model.ID != "" {
			models = append(models, model.ID)
		}
	}

	sort.Strings(models)

	return models, nil
}

// chatOnce performs a single chat completion request.
func chatOnce(
	ctx context.Context,
	provider providerInfo,
	model string,
	messages []apiMessage,
	schemas []map[string]any,
) (apiMessage, error) {
	payload := map[string]any{
		"model":    model,
		"messages": messages,
	}

	if len(schemas) > 0 {
		payload["tools"] = schemas
		payload["tool_choice"] = "auto"
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return apiMessage{}, fmt.Errorf("encode ai request: %w", err)
	}

	url := strings.TrimSuffix(provider.baseURL, "/") + "/chat/completions"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)

	if err != nil {
		return apiMessage{}, fmt.Errorf("build ai request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.apiKey)

	res, err := client.Do(req)

	if err != nil {
		return apiMessage{}, fmt.Errorf("ai request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))

	if err != nil {
		return apiMessage{}, fmt.Errorf("read ai response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return apiMessage{}, fmt.Errorf(
			"ai request failed (%d): %s",
			res.StatusCode,
			truncate(strings.TrimSpace(string(raw)), 300),
		)
	}

	var parsed struct {
		Choices []struct {
			Message apiMessage `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(raw, &parsed); err != nil {
		return apiMessage{}, fmt.Errorf("decode ai response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return apiMessage{}, errors.New("ai returned no choices")
	}

	return parsed.Choices[0].Message, nil
}

// executeToolCall runs one tool call inside the project root and
// returns its display label and output.
func executeToolCall(
	ctx context.Context,
	projectRoot string,
	call ToolCall,
) (display string, result string) {
	display = toolDisplay(call)

	output, err := tools.Run(
		ctx,
		projectRoot,
		call.Function.Name,
		json.RawMessage(call.Function.Arguments),
	)

	if err != nil {
		return display, "error: " + err.Error()
	}

	return display, truncate(output, maxToolResultBytes)
}

// toolDisplay builds the short label stored alongside a tool result,
// e.g. "Read File: src/main.go".
func toolDisplay(call ToolCall) string {
	var parsed struct {
		Path string `json:"path"`
	}

	_ = json.Unmarshal([]byte(call.Function.Arguments), &parsed)

	switch call.Function.Name {
	case "read_file":
		return "Read File: " + parsed.Path
	case "write_file":
		return "Write File: " + parsed.Path
	case "list_files":
		if parsed.Path == "" {
			return "List Files: ."
		}

		return "List Files: " + parsed.Path
	}

	return "Tool: " + call.Function.Name
}

// connectProvider loads a connected provider and decrypts its API key.
func connectProvider(
	ctx context.Context,
	providerName string,
) (providerInfo, error) {
	name, baseURL, encryptedKey, err := loadProvider(ctx, providerName)

	if err != nil {
		return providerInfo{}, err
	}

	apiKey, err := security.DecryptAPIKey(encryptedKey)

	if err != nil {
		return providerInfo{}, fmt.Errorf("decrypt api key: %w", err)
	}

	return providerInfo{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
	}, nil
}

// loadProvider returns the name, base URL and encrypted API key of
// the requested provider. With an empty name the first connected
// provider is used.
func loadProvider(
	ctx context.Context,
	providerName string,
) (name string, baseURL string, encryptedKey string, err error) {
	if providerName == "" {
		err = store.DB.QueryRowContext(ctx, `
			SELECT name, base_url, api_key FROM providers ORDER BY id LIMIT 1
		`).Scan(&name, &baseURL, &encryptedKey)
	} else {
		name = providerName

		err = store.DB.QueryRowContext(ctx, `
			SELECT base_url, api_key FROM providers WHERE name = ? LIMIT 1
		`, providerName).Scan(&baseURL, &encryptedKey)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", ErrNoProvider
	}

	if err != nil {
		return "", "", "", fmt.Errorf("load provider: %w", err)
	}

	return name, baseURL, encryptedKey, nil
}

func truncate(text string, limit int) string {
	if len(text) <= limit {
		return text
	}

	return text[:limit] + "..."
}
