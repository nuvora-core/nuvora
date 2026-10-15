package models

// Entry is a file or folder returned by the project file explorer.
type Entry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}
