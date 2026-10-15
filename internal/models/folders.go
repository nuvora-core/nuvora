package models

// Folder is a directory entry returned by the folder browser.
type Folder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
