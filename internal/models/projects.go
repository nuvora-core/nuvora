package models

type NewProjectReq struct {
	Name      string `json:"name"`
	RootPath  string `json:"root_path"`
	CreatedAt string `json:"created_at"`
}

// Project is a stored workspace project.
type Project struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	RootPath  string `json:"root_path"`
	CreatedAt string `json:"created_at"`
}
