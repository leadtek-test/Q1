package factory

import "testing"

func TestSingleton(t *testing.T) {
	callCount := 0
	s := NewSingleton(func(key string) any {
		callCount++
		return key + "-value"
	})

	v1 := s.Get("x").(string)
	v2 := s.Get("x").(string)
	if v1 != "x-value" || v2 != "x-value" {
		t.Fatalf("unexpected values: %s %s", v1, v2)
	}
	if callCount != 1 {
		t.Fatalf("supplier should be called once, got %d", callCount)
	}
}
