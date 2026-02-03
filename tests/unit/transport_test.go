package unit

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/internal/transport"
	"golang.org/x/time/rate"
)

// TestBodyCapEnforcement verifies that responses larger than MaxBodyBytes are handled correcty.
func TestBodyCapEnforcement(t *testing.T) {
	// 1. Setup server sending 1024 bytes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 1024)))
	}))
	defer ts.Close()

	// 2. Configure transport with 512 byte cap
	p := transport.DefaultPolicy()
	p.MaxBodyBytes = 512
	// Disable rate limit for this test
	p.RateLimiter = rate.NewLimiter(rate.Inf, 0)

	tr := transport.NewTransport(http.DefaultClient, p)

	// 3. Make request
	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	defer resp.Body.Close()

	// 4. ReadAll should read up to limit + 1
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// 5. Verify length.
	// We expect 513 bytes (512 allowed + 1 to prove overflow)
	expected := int(p.MaxBodyBytes + 1)
	if len(data) > expected {
		t.Errorf("Expected max %d bytes, got %d", expected, len(data))
	}
}

// TestRateLimiterBehavior verifies the transport blocks when rate limit is exceeded.
func TestRateLimiterBehavior(t *testing.T) {
	p := transport.DefaultPolicy()
	// Allow 1 request per second, burst 1
	p.RateLimiter = rate.NewLimiter(rate.Limit(1), 1)

	tr := transport.NewTransport(http.DefaultClient, p)

	// Server that does nothing
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	ctx := context.Background()
	req, _ := http.NewRequest("GET", ts.URL, nil)

	// First request should be instant (burst 1)
	start := time.Now()
	if _, err := tr.Do(ctx, req); err != nil {
		t.Fatalf("Req 1 failed: %v", err)
	}

	// Second request should block for ~1s
	if _, err := tr.Do(ctx, req); err != nil {
		t.Fatalf("Req 2 failed: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 900*time.Millisecond {
		t.Errorf("Rate limiter did not block long enough. Elapsed: %v", elapsed)
	}
}

// TestContextDeadline verifies the context is respected.
func TestContextDeadline(t *testing.T) {
	// Server sleeps longer than timeout
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer ts.Close()

	p := transport.DefaultPolicy()
	p.Timeout = 50 * time.Millisecond // fast timeout
	p.RateLimiter = rate.NewLimiter(rate.Inf, 0)

	tr := transport.NewTransport(http.DefaultClient, p)

	ctx := context.Background()
	req, _ := http.NewRequest("GET", ts.URL, nil)

	_, err := tr.Do(ctx, req)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
