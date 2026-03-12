package util

import "testing"

func TestMarshalString(t *testing.T) {
	s, err := MarshalString(map[string]int{"x": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != `{"x":1}` {
		t.Fatalf("unexpected json: %s", s)
	}
}
