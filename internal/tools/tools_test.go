package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRejectsPathsOutsideRoot(t *testing.T) {
	root := t.TempDir()

	if _, err := Resolve(root, "../outside.txt"); err == nil {
		t.Error("Resolve() allowed ../outside.txt")
	}

	if _, err := Resolve(root, "/etc/passwd"); err == nil {
		t.Error("Resolve() allowed an absolute path outside the root")
	}

	got, err := Resolve(root, "sub/dir/file.txt")

	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	want := filepath.Join(root, "sub", "dir", "file.txt")

	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks are not supported:", err)
	}

	if _, err := Resolve(root, "link/secret.txt"); err == nil {
		t.Error("Resolve() allowed a symlink pointing outside the root")
	}
}

func TestWriteReadListRoundTrip(t *testing.T) {
	root := t.TempDir()
	ctx := context.Background()

	args, err := json.Marshal(map[string]string{
		"path":    "notes/hello.txt",
		"content": "merhaba",
	})

	if err != nil {
		t.Fatal(err)
	}

	if _, err := Run(ctx, root, "write_file", args); err != nil {
		t.Fatalf("write_file error: %v", err)
	}

	args, err = json.Marshal(map[string]string{"path": "notes/hello.txt"})

	if err != nil {
		t.Fatal(err)
	}

	content, err := Run(ctx, root, "read_file", args)

	if err != nil {
		t.Fatalf("read_file error: %v", err)
	}

	if content != "merhaba" {
		t.Errorf("read_file = %q, want %q", content, "merhaba")
	}

	args, err = json.Marshal(map[string]string{"path": "notes"})

	if err != nil {
		t.Fatal(err)
	}

	listing, err := Run(ctx, root, "list_files", args)

	if err != nil {
		t.Fatalf("list_files error: %v", err)
	}

	if !strings.Contains(listing, "hello.txt") {
		t.Errorf("list_files = %q, want it to contain hello.txt", listing)
	}

	// Traversal is rejected through the tool layer too.
	args, err = json.Marshal(map[string]string{
		"path":    "../escape.txt",
		"content": "x",
	})

	if err != nil {
		t.Fatal(err)
	}

	if _, err := Run(ctx, root, "write_file", args); err == nil {
		t.Error("write_file allowed a path outside the root")
	}
}
