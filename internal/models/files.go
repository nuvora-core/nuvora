package models

// SaveFileReq is the request body for saving a file.
type SaveFileReq struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
