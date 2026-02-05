package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/internal/transport"
)

// NOTE: We cannot test unexported functions like `shouldRetry` or `isIdempotent` directly from external package.
// We must test them via public API (`DoJSON`, `Do`) or export them for testing (not recommended).

//
// ISSUE: `shouldRetry` and `isIdempotent` are unexported in `internal/transport`.
// If I move the tests to `tests/unit`, I cannot call `transport.shouldRetry`.
// I will have to adapt the tests to test the *public surface* (DoJSON) behavior that relies on them.

// TestIdempotencyBehavior verifies that unsafe methods are NOT retried.
func TestIdempotencyBehavior(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	}))
	defer ts.Close()

	p := transport.DefaultPolicy()
	p.Retry.BaseDelay = 1 * time.Millisecond

	tr := transport.NewTransport(http.DefaultClient, p)

	// POST is not idempotent -> should fail on logic error (500) without retry
	_, err := tr.DoJSON(context.Background(), "POST", ts.URL, nil, nil)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if calls != 1 {
		t.Errorf("Expected 1 call (no retry), got %d", calls)
	}
}

// TestDoJSONRetryLimit verifies we don't retry forever.
func TestDoJSONRetryLimit(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	}))
	defer ts.Close()

	p := transport.DefaultPolicy()
	p.Retry.MaxRetries = 2
	p.Retry.BaseDelay = 1 * time.Millisecond // fast test
	p.Retry.MaxDelay = 5 * time.Millisecond

	tr := transport.NewTransport(http.DefaultClient, p)

	// GET is idempotent -> should retry
	_, err := tr.DoJSON(context.Background(), "GET", ts.URL, nil, nil)

	if err == nil {
		t.Error("Expected error after retries exhausted, got nil")
	}

	// Initial call (1) + 2 retries = 3 calls total
	expectedCalls := 1 + p.Retry.MaxRetries
	if calls != expectedCalls {
		t.Errorf("Expected %d calls, got %d", expectedCalls, calls)
	}
}

// TestDoJSONNonIdempotentNoRetry verifies POST is not retried on logic errors (like 500).
func TestDoJSONNonIdempotentNoRetry(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	}))
	defer ts.Close()

	p := transport.DefaultPolicy()

	p.Retry.BaseDelay = 1 * time.Millisecond

	tr := transport.NewTransport(http.DefaultClient, p)

	// POST -> should NOT retry on 500
	_, err := tr.DoJSON(context.Background(), "POST", ts.URL, nil, nil)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if calls != 1 {
		t.Errorf("Expected exactly 1 call for POST 500, got %d", calls)
	}
}
