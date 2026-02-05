package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/types"
)

func TestNon2xxReturnsTypedStatusAndKind(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"message":"rate limited"}`))
	}))
	defer srv.Close()

	p := DefaultPolicy()
	p.Retry.MaxRetries = 0 // no retries for this test
	tr := NewTransport(http.DefaultClient, p)

	_, err := tr.DoJSON(context.Background(), http.MethodGet, srv.URL+"/x", nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Assert errors.As(err, *types.Status)
	var st *types.Status
	if !errors.As(err, &st) {
		t.Fatalf("expected errors.As to find *types.Status, got: %T %v", err, err)
	}
	if st.StatusCode != 429 {
		t.Fatalf("expected status 429, got %d", st.StatusCode)
	}
	if st.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %s", st.Method)
	}
	if st.Path != "/x" {
		t.Fatalf("expected path /x, got %s", st.Path)
	}

	// Assert KindStatus
	var top *types.Error
	if !errors.As(err, &top) {
		t.Fatalf("expected errors.As to find *types.Error, got: %T %v", err, err)
	}
	if top.Kind() != types.KindStatus {
		t.Fatalf("expected kind %s, got %s", types.KindStatus, top.Kind())
	}
}

func TestRetryReplaysBodyOnRetryForPUT(t *testing.T) {
	var n int32
	var firstBody, secondBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&n, 1)

		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		if count == 1 {
			firstBody = b
			w.WriteHeader(500) // transient -> should retry for PUT
			_, _ = w.Write([]byte("fail"))
			return
		}

		secondBody = b
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	p := DefaultPolicy()
	p.Timeout = 2 * time.Second
	p.Retry.MaxRetries = 2
	p.Retry.BaseDelay = 1 * time.Millisecond
	p.Retry.MaxDelay = 5 * time.Millisecond

	tr := NewTransport(http.DefaultClient, p)

	out, err := tr.DoJSON(context.Background(), http.MethodPut, srv.URL+"/retry", nil, []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if string(out) != "ok" {
		t.Fatalf("expected ok, got %q", string(out))
	}
	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("expected 2 attempts, got %d", n)
	}
	if string(firstBody) != `{"hello":"world"}` {
		t.Fatalf("first body mismatch: %q", string(firstBody))
	}
	if string(secondBody) != `{"hello":"world"}` {
		t.Fatalf("second body mismatch (replay failed): %q", string(secondBody))
	}
}
