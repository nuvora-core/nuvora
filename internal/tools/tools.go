// Package tools provides the functions the AI can call to work with
// files inside the active project. Every tool is scoped to the
// project root_path; paths passed by the model are relative to it and
// can never escape it.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool is a function the AI can call inside a project.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Execute     func(ctx context.Context, projectRoot string, args json.RawMessage) (string, error)
}

var registry = []Tool{
	readFileTool,
	writeFileTool,
	listFilesTool,
}

// All returns the registered tools.
func All() []Tool {
	return registry
}

// Find returns the tool with the given name.
func Find(name string) (Tool, bool) {
	for _, tool := range registry {
		if tool.Name == name {
			return tool, true
		}
	}

	return Tool{}, false
}

// Run executes a tool by name within the given project root.
func Run(
	ctx context.Context,
	projectRoot string,
	name string,
	args json.RawMessage,
) (string, error) {
	tool, ok := Find(name)

	if !ok {
		return "", fmt.Errorf("unknown tool %q", name)
	}

	return tool.Execute(ctx, projectRoot, args)
}

// Schemas returns the tools in OpenAI function calling format.
func Schemas() []map[string]any {
	schemas := make([]map[string]any, 0, len(registry))

	for _, tool := range registry {
		schemas = append(schemas, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}

	return schemas
}
