package config

import "testing"

func TestGetRelativePathFromCaller(t *testing.T) {
	rel, err := getRelativePathFromCaller()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel == "" {
		t.Fatalf("relative path should not be empty")
	}
}
