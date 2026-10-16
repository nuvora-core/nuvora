package sessions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nuvora-core/nuvora/internal/ai"
	"github.com/nuvora-core/nuvora/internal/models"
	"github.com/nuvora-core/nuvora/internal/store"
)

// maxMessageLength caps the length of a single chat message.
const maxMessageLength = 16384

// CreateSession creates a new chat session for a project.
func CreateSession(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req models.NewSessionReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid request",
		})

		return
	}

	if req.ProjectID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "project id is required",
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

	var projectCount int

	err = store.DB.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM projects WHERE id = ?
	`, req.ProjectID).Scan(&projectCount)

	if err != nil {
		fmt.Println("sessions: count projects error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database error",
		})

		return
	}

	if projectCount == 0 {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "project not found",
		})

		return
	}

	result, err := store.DB.ExecContext(r.Context(), `
		INSERT INTO sessions (project_id, title) VALUES (?, ?)
	`, req.ProjectID, strings.TrimSpace(req.Title))

	if err != nil {
		fmt.Println("sessions: create session error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to create session",
		})

		return
	}

	id, err := result.LastInsertId()

	if err != nil {
		fmt.Println("sessions: create session id error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to create session",
		})

		return
	}

	session, err := loadSession(r.Context(), int(id))

	if err != nil {
		fmt.Println("sessions: load session error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to create session",
		})

		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]any{
		"session": session,
	})
}

// ListSessions lists the sessions of a project, newest first.
func ListSessions(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	projectID, err := strconv.Atoi(r.URL.Query().Get("project_id"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid project id",
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

	rows, err := store.DB.QueryContext(r.Context(), `
		SELECT id, project_id, title, created_at
		FROM sessions
		WHERE project_id = ?
		ORDER BY id DESC
	`, projectID)

	if err != nil {
		fmt.Println("sessions: list sessions error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load sessions",
		})

		return
	}

	defer rows.Close()

	sessions := []models.Session{}

	for rows.Next() {
		var session models.Session

		if err := rows.Scan(
			&session.ID,
			&session.ProjectID,
			&session.Title,
			&session.CreatedAt,
		); err != nil {
			fmt.Println("sessions: scan session error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "failed to load sessions",
			})

			return
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("sessions: iterate sessions error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load sessions",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"sessions": sessions,
	})
}

// GetMessages returns a session with its full message history.
func GetMessages(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	sessionID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid session id",
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

	session, err := loadSession(r.Context(), sessionID)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "session not found",
		})

		return
	}

	if err != nil {
		fmt.Println("sessions: load session error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database error",
		})

		return
	}

	messages, err := loadMessages(r.Context(), sessionID)

	if err != nil {
		fmt.Println("sessions: load messages error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load messages",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"session":  session,
		"messages": messages,
	})
}

// SendMessage stores the user message, sends the full history to the
// AI and stores the assistant reply.
func SendMessage(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	sessionID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid session id",
		})

		return
	}

	var req models.SendMessageReq

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid request",
		})

		return
	}

	content := strings.TrimSpace(req.Content)

	if content == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "content is required",
		})

		return
	}

	if len(content) > maxMessageLength {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "message is too long",
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

	session, err := loadSession(r.Context(), sessionID)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "session not found",
		})

		return
	}

	if err != nil {
		fmt.Println("sessions: load session error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database error",
		})

		return
	}

	// The user message is stored first so nothing is lost when the
	// AI request fails.
	userResult, err := store.DB.ExecContext(r.Context(), `
		INSERT INTO messages (session_id, role, content)
		VALUES (?, 'user', ?)
	`, sessionID, content)

	if err != nil {
		fmt.Println("sessions: save user message error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to save message",
		})

		return
	}

	userID, err := userResult.LastInsertId()

	if err != nil {
		fmt.Println("sessions: save user message id error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to save message",
		})

		return
	}

	// The first message names the session.
	title := session.Title

	if title == "" {
		title = excerpt(content, 60)

		_, err = store.DB.ExecContext(r.Context(), `
			UPDATE sessions SET title = ? WHERE id = ?
		`, title, sessionID)

		if err != nil {
			fmt.Println("sessions: update session title error:", err)
		}
	}

	historyRows, err := loadMessages(r.Context(), sessionID)

	if err != nil {
		fmt.Println("sessions: load history error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load message history",
		})

		return
	}

	var rootPath string

	err = store.DB.QueryRowContext(r.Context(), `
		SELECT root_path FROM projects WHERE id = ?
	`, session.ProjectID).Scan(&rootPath)

	if err != nil {
		fmt.Println("sessions: load project error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database error",
		})

		return
	}

	history := make([]ai.Message, 0, len(historyRows))

	for _, message := range historyRows {
		role := message.Role

		// Tool lines are replayed as assistant context for the model.
		if role == "tool" {
			role = "assistant"
		}

		history = append(history, ai.Message{
			Role:    role,
			Content: message.Content,
		})
	}

	result, err := ai.ChatWithTools(r.Context(), req.Provider, req.Model, history, rootPath)

	if errors.Is(err, ai.ErrNoProvider) {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "no provider connected",
		})

		return
	}

	if err != nil {
		fmt.Println("sessions: ai chat error:", err)

		w.WriteHeader(http.StatusBadGateway)

		json.NewEncoder(w).Encode(map[string]any{
			"error":  "ai request failed",
			"reason": err.Error(),
		})

		return
	}

	toolMessages := []models.ChatMessage{}

	for _, event := range result.ToolEvents {
		toolResult, err := store.DB.ExecContext(r.Context(), `
			INSERT INTO messages (session_id, role, content)
			VALUES (?, 'tool', ?)
		`, sessionID, event.Display+"\n\n"+event.Result)

		if err != nil {
			fmt.Println("sessions: save tool message error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "failed to save tool output",
			})

			return
		}

		toolID, err := toolResult.LastInsertId()

		if err != nil {
			fmt.Println("sessions: save tool message id error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "failed to save tool output",
			})

			return
		}

		toolMessage, err := messageByID(r.Context(), toolID)

		if err != nil {
			fmt.Println("sessions: load tool message error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "failed to load tool output",
			})

			return
		}

		toolMessages = append(toolMessages, toolMessage)
	}

	replyResult, err := store.DB.ExecContext(r.Context(), `
		INSERT INTO messages (session_id, role, content)
		VALUES (?, 'assistant', ?)
	`, sessionID, result.Reply)

	if err != nil {
		fmt.Println("sessions: save assistant message error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to save reply",
		})

		return
	}

	replyID, err := replyResult.LastInsertId()

	if err != nil {
		fmt.Println("sessions: save assistant message id error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to save reply",
		})

		return
	}

	userMessage, err := messageByID(r.Context(), userID)

	if err != nil {
		fmt.Println("sessions: load user message error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load message",
		})

		return
	}

	assistantMessage, err := messageByID(r.Context(), replyID)

	if err != nil {
		fmt.Println("sessions: load assistant message error:", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to load message",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"title":             title,
		"user_message":      userMessage,
		"tool_messages":     toolMessages,
		"assistant_message": assistantMessage,
	})
}

func loadSession(
	ctx context.Context,
	id int,
) (models.Session, error) {
	var session models.Session

	err := store.DB.QueryRowContext(ctx, `
		SELECT id, project_id, title, created_at FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID,
		&session.ProjectID,
		&session.Title,
		&session.CreatedAt,
	)

	return session, err
}

func loadMessages(
	ctx context.Context,
	sessionID int,
) ([]models.ChatMessage, error) {
	rows, err := store.DB.QueryContext(ctx, `
		SELECT id, session_id, role, content, created_at
		FROM messages
		WHERE session_id = ?
		ORDER BY id
	`, sessionID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	messages := []models.ChatMessage{}

	for rows.Next() {
		var message models.ChatMessage

		if err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func messageByID(
	ctx context.Context,
	id int64,
) (models.ChatMessage, error) {
	var message models.ChatMessage

	err := store.DB.QueryRowContext(ctx, `
		SELECT id, session_id, role, content, created_at FROM messages WHERE id = ?
	`, id).Scan(
		&message.ID,
		&message.SessionID,
		&message.Role,
		&message.Content,
		&message.CreatedAt,
	)

	return message, err
}

// excerpt shortens a message into a session title.
func excerpt(text string, limit int) string {
	runes := []rune(text)

	if len(runes) <= limit {
		return text
	}

	return strings.TrimSpace(string(runes[:limit])) + "..."
}
