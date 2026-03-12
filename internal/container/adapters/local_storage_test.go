package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceRepositoryLocal(t *testing.T) {
	root := t.TempDir()
	repo := NewWorkspaceRepositoryLocal(root)

	dir, err := repo.EnsureUserDir(12)
	if err != nil {
		t.Fatalf("EnsureUserDir failed: %v", err)
	}
	if _, statErr := os.Stat(dir); statErr != nil {
		t.Fatalf("user dir not created: %v", statErr)
	}

	path, err := repo.Save(12, "../a.txt", []byte("abc"))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if filepath.Base(path) != "a.txt" {
		t.Fatalf("unexpected saved file name: %s", path)
	}
}

func TestObjectStorageRepositoryLocal(t *testing.T) {
	root := t.TempDir()
	repo := NewObjectStorageRepositoryLocal(root)

	err := repo.Upload(context.Background(), "../data/object.txt", strings.NewReader("hello"), 5, "text/plain")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	target := filepath.Join(root, "data", "object.txt")
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected content: %s", string(content))
	}
}
