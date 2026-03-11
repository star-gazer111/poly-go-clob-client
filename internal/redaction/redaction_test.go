package redaction

import (
	"net/http"
	"testing"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"   ", ""},
		{"a", "***"},
		{"abcd", "***"},
		{"abcdefgh", "***"},               // <= 8 -> fully redacted
		{"abcdefghi", "abc***ghi"},        // 9 -> partial
		{"0123456789abcdef", "012***def"}, // long
	}

	for _, tt := range tests {
		got := Redact(tt.in)
		if got != tt.want {
			t.Fatalf("Redact(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestRedactHeaders_DoesNotMutateInput(t *testing.T) {
	h := make(http.Header)
	h.Set("Authorization", "Bearer supersecrettoken")
	h.Add("X-Api-Key", "myapikey")
	h.Set("User-Agent", "ua")
	h.Set("X-Custom", "somevalue")

	out := RedactHeaders(h)

	// Original must remain unchanged
	if h.Get("Authorization") != "Bearer supersecrettoken" {
		t.Fatalf("input mutated: Authorization=%q", h.Get("Authorization"))
	}
	if h.Get("X-Api-Key") != "myapikey" {
		t.Fatalf("input mutated: X-Api-Key=%q", h.Get("X-Api-Key"))
	}

	// Output should be redacted
	if out.Get("Authorization") != "***" && out.Get("Authorization") != "Bearer ***" {
		// Our RedactHeaders uses name-based redaction => "***" for Authorization
		// but heuristic may also apply. Accept either safe outcome.
		t.Fatalf("expected Authorization redacted, got=%q", out.Get("Authorization"))
	}
	if out.Get("X-Api-Key") != "***" {
		t.Fatalf("expected X-Api-Key redacted, got=%q", out.Get("X-Api-Key"))
	}

	// Non-sensitive should pass through
	if out.Get("User-Agent") != "ua" {
		t.Fatalf("expected User-Agent preserved, got=%q", out.Get("User-Agent"))
	}
	if out.Get("X-Custom") != "somevalue" {
		t.Fatalf("expected X-Custom preserved, got=%q", out.Get("X-Custom"))
	}
}

func TestRedactHeaders_BearerHeuristic(t *testing.T) {
	h := make(http.Header)
	h.Set("X-Whatever", "Bearer abcdefghijklmnopqrstuvwxyz")

	out := RedactHeaders(h)
	if out.Get("X-Whatever") != "Bearer ***" {
		t.Fatalf("expected Bearer token redacted, got=%q", out.Get("X-Whatever"))
	}
}
