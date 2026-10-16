package models

type NewSessionReq struct {
	ProjectID int    `json:"project_id"`
	Title     string `json:"title"`
}

// Session is a chat session that belongs to a single project.
type Session struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

type SendMessageReq struct {
	Content  string `json:"content"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// ChatMessage is a single message in a session. Role is either
// "user" or "assistant".
type ChatMessage struct {
	ID        int    `json:"id"`
	SessionID int    `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
