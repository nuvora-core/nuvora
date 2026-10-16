package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/nuvora-core/nuvora/internal/ai"
	"github.com/nuvora-core/nuvora/internal/config"
	"github.com/nuvora-core/nuvora/internal/models"
	"github.com/nuvora-core/nuvora/internal/security"
	"github.com/nuvora-core/nuvora/internal/store"
)

func perform(
	t *testing.T,
	handler http.HandlerFunc,
	method string,
	body string,
	params map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader

	if body != "" {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, "/", reader)

	if len(params) > 0 {
		rctx := chi.NewRouteContext()

		for key, value := range params {
			rctx.URLParams.Add(key, value)
		}

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}

func TestSendMessageSendsFullHistory(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := config.Init(); err != nil {
		t.Fatalf("config.Init() error: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := store.ConnectDB(dbPath); err != nil {
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

	// Fake provider that records every request it receives.
	type chatRequest struct {
		Model    string       `json:"model"`
		Messages []ai.Message `json:"messages"`
	}

	requests := []chatRequest{}

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode ai request: %v", err)
		}

		requests = append(requests, req)

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "hello from ai"}},
			},
		})
	}))

	t.Cleanup(aiServer.Close)

	encryptedKey, err := security.EncryptAPIKey("test-key")

	if err != nil {
		t.Fatalf("security.EncryptAPIKey() error: %v", err)
	}

	_, err = store.DB.Exec(`
		INSERT INTO providers (name, base_url, api_key) VALUES ('groq', ?, ?)
	`, aiServer.URL, encryptedKey)

	if err != nil {
		t.Fatalf("insert provider error: %v", err)
	}

	_, err = store.DB.Exec(`
		INSERT INTO projects (name, root_path) VALUES ('demo', '/tmp')
	`)

	if err != nil {
		t.Fatalf("insert project error: %v", err)
	}

	// Create a session for the project.
	rec := perform(t, CreateSession, http.MethodPost, `{"project_id": 1}`, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateSession status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	var createRes struct {
		Session models.Session `json:"session"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &createRes); err != nil {
		t.Fatalf("decode CreateSession response: %v", err)
	}

	sessionID := strconv.Itoa(createRes.Session.ID)
	params := map[string]string{"id": sessionID}

	// First message.
	rec = perform(t, SendMessage, http.MethodPost, `{"content": "Hi there"}`, params)

	if rec.Code != http.StatusOK {
		t.Fatalf("SendMessage #1 status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body)
	}

	var firstRes struct {
		Title            string             `json:"title"`
		AssistantMessage models.ChatMessage `json:"assistant_message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &firstRes); err != nil {
		t.Fatalf("decode SendMessage #1 response: %v", err)
	}

	if firstRes.AssistantMessage.Content != "hello from ai" {
		t.Errorf("assistant content = %q, want %q", firstRes.AssistantMessage.Content, "hello from ai")
	}

	if firstRes.Title != "Hi there" {
		t.Errorf("session title = %q, want %q", firstRes.Title, "Hi there")
	}

	// Second message must carry the whole history to the AI.
	rec = perform(t, SendMessage, http.MethodPost, `{"content": "How are you?"}`, params)

	if rec.Code != http.StatusOK {
		t.Fatalf("SendMessage #2 status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body)
	}

	if len(requests) != 2 {
		t.Fatalf("ai request count = %d, want 2", len(requests))
	}

	if requests[0].Model != "llama-3.3-70b-versatile" {
		t.Errorf("model = %q, want default groq model", requests[0].Model)
	}

	wantHistory := []ai.Message{
		{Role: "user", Content: "Hi there"},
		{Role: "assistant", Content: "hello from ai"},
		{Role: "user", Content: "How are you?"},
	}

	if !reflect.DeepEqual(requests[1].Messages, wantHistory) {
		t.Errorf("second ai request history = %+v, want %+v", requests[1].Messages, wantHistory)
	}

	// The stored history contains both turns.
	rec = perform(t, GetMessages, http.MethodGet, "", params)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetMessages status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body)
	}

	var messagesRes struct {
		Messages []models.ChatMessage `json:"messages"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &messagesRes); err != nil {
		t.Fatalf("decode GetMessages response: %v", err)
	}

	if len(messagesRes.Messages) != 4 {
		t.Errorf("stored message count = %d, want 4", len(messagesRes.Messages))
	}
}

func TestSendMessageRunsToolCalls(t *testing.T) {
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

	projectRoot := t.TempDir()

	if err := os.WriteFile(filepath.Join(projectRoot, "hello.txt"), []byte("merhaba dunya"), 0o644); err != nil {
		t.Fatalf("write project file error: %v", err)
	}

	// First AI call requests a read_file tool call, the second one
	// gives the final answer.
	type capturedRequest struct {
		Messages []struct {
			Role       string `json:"role"`
			Content    string `json:"content"`
			ToolCallID string `json:"tool_call_id"`
		} `json:"messages"`
	}

	requests := []capturedRequest{}
	call := 0

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req capturedRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode ai request: %v", err)
		}

		requests = append(requests, req)

		w.Header().Set("Content-Type", "application/json")

		if call == 0 {
			call++

			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{
						"message": map[string]any{
							"role":    "assistant",
							"content": "",
							"tool_calls": []map[string]any{
								{
									"id":   "call_1",
									"type": "function",
									"function": map[string]any{
										"name":      "read_file",
										"arguments": `{"path": "hello.txt"}`,
									},
								},
							},
						},
					},
				},
			})

			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "the file says merhaba"}},
			},
		})
	}))

	t.Cleanup(aiServer.Close)

	encryptedKey, err := security.EncryptAPIKey("test-key")

	if err != nil {
		t.Fatalf("security.EncryptAPIKey() error: %v", err)
	}

	_, err = store.DB.Exec(`
		INSERT INTO providers (name, base_url, api_key) VALUES ('groq', ?, ?)
	`, aiServer.URL, encryptedKey)

	if err != nil {
		t.Fatalf("insert provider error: %v", err)
	}

	_, err = store.DB.Exec(`
		INSERT INTO projects (name, root_path) VALUES ('demo', ?)
	`, projectRoot)

	if err != nil {
		t.Fatalf("insert project error: %v", err)
	}

	rec := perform(t, CreateSession, http.MethodPost, `{"project_id": 1}`, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateSession status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	var createRes struct {
		Session models.Session `json:"session"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &createRes); err != nil {
		t.Fatalf("decode CreateSession response: %v", err)
	}

	params := map[string]string{"id": strconv.Itoa(createRes.Session.ID)}

	rec = perform(t, SendMessage, http.MethodPost, `{"content": "what is in hello.txt?"}`, params)

	if rec.Code != http.StatusOK {
		t.Fatalf("SendMessage status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body)
	}

	var sendRes struct {
		ToolMessages     []models.ChatMessage `json:"tool_messages"`
		AssistantMessage models.ChatMessage   `json:"assistant_message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &sendRes); err != nil {
		t.Fatalf("decode SendMessage response: %v", err)
	}

	if sendRes.AssistantMessage.Content != "the file says merhaba" {
		t.Errorf("assistant content = %q, want the final answer", sendRes.AssistantMessage.Content)
	}

	if len(sendRes.ToolMessages) != 1 {
		t.Fatalf("tool message count = %d, want 1", len(sendRes.ToolMessages))
	}

	toolMessage := sendRes.ToolMessages[0]

	if toolMessage.Role != "tool" {
		t.Errorf("tool message role = %q, want %q", toolMessage.Role, "tool")
	}

	if !strings.HasPrefix(toolMessage.Content, "Read File: hello.txt") {
		t.Errorf("tool message = %q, want it to start with %q", toolMessage.Content, "Read File: hello.txt")
	}

	// The second AI call must carry the tool result back.
	if len(requests) != 2 {
		t.Fatalf("ai request count = %d, want 2", len(requests))
	}

	found := false

	for _, message := range requests[1].Messages {
		if message.Role == "tool" && message.ToolCallID == "call_1" && strings.Contains(message.Content, "merhaba dunya") {
			found = true
		}
	}

	if !found {
		t.Errorf("second ai request does not contain the tool result: %+v", requests[1].Messages)
	}
}
