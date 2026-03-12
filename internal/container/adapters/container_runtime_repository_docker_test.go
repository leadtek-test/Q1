package adapters

import "testing"

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
