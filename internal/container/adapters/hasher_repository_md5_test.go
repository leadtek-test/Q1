package adapters

import "testing"

func TestNewHasherRepositoryMD5(t *testing.T) {
	hasher := NewHasherRepositoryMD5()
	if hasher == nil {
		t.Fatal("expected hasher instance, got nil")
	}
}

func TestHasherRepositoryMD5Hash(t *testing.T) {
	hasher := NewHasherRepositoryMD5()

	cases := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "empty string",
			raw:      "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "normal string",
			raw:      "leadtek",
			expected: "ad75acd07c907a115b75bf1934e17d40",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasher.Hash(tc.raw); got != tc.expected {
				t.Fatalf("unexpected hash: got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestHasherRepositoryMD5Compare(t *testing.T) {
	hasher := NewHasherRepositoryMD5()

	t.Run("match", func(t *testing.T) {
		encoded := hasher.Hash("leadtek")
		if !hasher.Compare("leadtek", encoded) {
			t.Fatal("expected compare to return true")
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		encoded := hasher.Hash("leadtek")
		if hasher.Compare("other", encoded) {
			t.Fatal("expected compare to return false")
		}
	})
}
