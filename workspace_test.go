package q1_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceModules(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	modules := []string{
		"internal/common",
		"internal/container",
	}

	for _, module := range modules {
		moduleDir := filepath.Join(root, module)
		cacheKey := strings.ReplaceAll(module, "/", "-")

		cmd := exec.Command("go", "test", "./...")
		cmd.Dir = moduleDir
		cmd.Env = append(os.Environ(), "GOCACHE=/tmp/go-build-"+cacheKey)

		output, runErr := cmd.CombinedOutput()
		if runErr != nil {
			t.Fatalf("module %s tests failed: %v\n%s", module, runErr, string(output))
		}
	}
}
