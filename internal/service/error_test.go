package service

import "testing"

func TestErrorMessageIncludesProblems(t *testing.T) {
	// A validation error carries no Message, only field problems; Error() must
	// still produce a non-empty, deterministic message.
	err := Validation(map[string]string{
		"name":    "cannot be blank",
		"payload": "is required",
	})

	want := "name: cannot be blank; payload: is required"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorMessagePlain(t *testing.T) {
	if got := NotFoundf("payload %q not found", "x").Error(); got != `payload "x" not found` {
		t.Errorf("unexpected: %q", got)
	}
}
