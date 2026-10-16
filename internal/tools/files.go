package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// maxFileSize caps the size of files the tools read or write.
const maxFileSize = 1 << 20 // 1 MiB

var readFileTool = Tool{
	Name:        "read_file",
	Description: "Read a text file inside the project. The path is relative to the project root.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path relative to the project root, e.g. \"src/main.go\"",
			},
		},
		"required": []string{"path"},
	},
	Execute: runReadFile,
}

var writeFileTool = Tool{
	Name:        "write_file",
	Description: "Create or overwrite a text file inside the project. The path is relative to the project root, missing parent folders are created.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path relative to the project root, e.g. \"src/main.go\"",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Full content of the file",
			},
		},
		"required": []string{"path", "content"},
	},
	Execute: runWriteFile,
}

var listFilesTool = Tool{
	Name:        "list_files",
	Description: "List files and folders inside a project directory. The path is relative to the project root and defaults to the root itself.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path relative to the project root, e.g. \"src\". Defaults to the root.",
			},
		},
		"required": []string{},
	},
	Execute: runListFiles,
}

type readFileArgs struct {
	Path string `json:"path"`
}

func runReadFile(
	ctx context.Context,
	projectRoot string,
	args json.RawMessage,
) (string, error) {
	var parsed readFileArgs

	if err := json.Unmarshal(args, &parsed); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if parsed.Path == "" {
		return "", errors.New("path is required")
	}

	abs, err := Resolve(projectRoot, parsed.Path)

	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)

	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%s is not a file", parsed.Path)
	}

	if info.Size() > maxFileSize {
		return "", fmt.Errorf("%s is too large to read", parsed.Path)
	}

	data, err := os.ReadFile(abs)

	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", parsed.Path, err)
	}

	if bytes.IndexByte(data, 0) != -1 {
		return "", fmt.Errorf("%s is a binary file", parsed.Path)
	}

	return string(data), nil
}

type writeFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func runWriteFile(
	ctx context.Context,
	projectRoot string,
	args json.RawMessage,
) (string, error) {
	var parsed writeFileArgs

	if err := json.Unmarshal(args, &parsed); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if parsed.Path == "" {
		return "", errors.New("path is required")
	}

	if len(parsed.Content) > maxFileSize {
		return "", errors.New("content is too large")
	}

	abs, err := Resolve(projectRoot, parsed.Path)

	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", fmt.Errorf("cannot create folders for %s: %w", parsed.Path, err)
	}

	if err := os.WriteFile(abs, []byte(parsed.Content), 0o644); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", parsed.Path, err)
	}

	return fmt.Sprintf("wrote %d bytes to %s", len(parsed.Content), parsed.Path), nil
}

type listFilesArgs struct {
	Path string `json:"path"`
}

func runListFiles(
	ctx context.Context,
	projectRoot string,
	args json.RawMessage,
) (string, error) {
	var parsed listFilesArgs

	if len(args) > 0 {
		if err := json.Unmarshal(args, &parsed); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}

	abs, err := Resolve(projectRoot, parsed.Path)

	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)

	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s is not a folder", displayPath(parsed.Path))
	}

	entries, err := os.ReadDir(abs)

	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", displayPath(parsed.Path), err)
	}

	lines := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			lines = append(lines, entry.Name()+"/")
		} else {
			lines = append(lines, entry.Name())
		}
	}

	sort.Strings(lines)

	if len(lines) == 0 {
		return "(empty)", nil
	}

	return strings.Join(lines, "\n"), nil
}

func displayPath(path string) string {
	if path == "" {
		return "."
	}

	return path
}
