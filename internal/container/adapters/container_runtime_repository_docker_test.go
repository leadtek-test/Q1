package adapters

import (
	"path/filepath"
	"testing"
)

func TestSanitizeRuntimeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "  My App  ", want: "my-app"},
		{in: "NICE_NAME_123", want: "nice-name-123"},
		{in: "----", want: "default"},
		{in: "", want: "default"},
		{in: "a", want: "a"},
	}

	for _, c := range cases {
		if got := sanitizeRuntimeName(c.in); got != c.want {
			t.Fatalf("sanitizeRuntimeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolveWorkspaceMountPath(t *testing.T) {
	t.Run("default mode resolves absolute path", func(t *testing.T) {
		root := t.TempDir()
		target := filepath.Join(root, "1")

		repo := ContainerRuntimeRepositoryDocker{
			workspaceRoot:        root,
			workspaceRuntimeRoot: "",
		}

		got, err := repo.resolveWorkspaceMountPath(target)
		if err != nil {
			t.Fatalf("resolveWorkspaceMountPath failed: %v", err)
		}

		want := filepath.Clean(target)
		if got != want {
			t.Fatalf("unexpected mount path: got=%q want=%q", got, want)
		}
	})

	t.Run("runtime root mapping mode", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		runtimeRoot := t.TempDir()
		target := filepath.Join(workspaceRoot, "1", "sub")

		repo := ContainerRuntimeRepositoryDocker{
			workspaceRoot:        workspaceRoot,
			workspaceRuntimeRoot: runtimeRoot,
		}

		got, err := repo.resolveWorkspaceMountPath(target)
		if err != nil {
			t.Fatalf("resolveWorkspaceMountPath failed: %v", err)
		}

		want := filepath.Join(runtimeRoot, "1", "sub")
		if got != want {
			t.Fatalf("unexpected mapped mount path: got=%q want=%q", got, want)
		}
	})

	t.Run("runtime root set but workspace root missing", func(t *testing.T) {
		repo := ContainerRuntimeRepositoryDocker{
			workspaceRoot:        "",
			workspaceRuntimeRoot: t.TempDir(),
		}

		_, err := repo.resolveWorkspaceMountPath("workspace/1")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("workspace path outside workspace root should fail", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		runtimeRoot := t.TempDir()
		outside := t.TempDir()

		repo := ContainerRuntimeRepositoryDocker{
			workspaceRoot:        workspaceRoot,
			workspaceRuntimeRoot: runtimeRoot,
		}

		_, err := repo.resolveWorkspaceMountPath(outside)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
