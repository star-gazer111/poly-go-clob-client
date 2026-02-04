package clob

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/types"
)

func TestNewPublicClient_BaseURLValidation(t *testing.T) {
	_, err := NewPublicClient("")
	if err == nil {
		t.Fatal("expected error for empty baseURL")
	}
	var te *types.Error
	if !errors.As(err, &te) || te.Kind() != types.KindValidation {
		t.Fatalf("expected KindValidation error, got: %T %v", err, err)
	}

	_, err = NewPublicClient("not a url")
	if err == nil {
		t.Fatal("expected error for invalid url")
	}
	if !errors.As(err, &te) || te.Kind() != types.KindValidation {
		t.Fatalf("expected KindValidation error, got: %T %v", err, err)
	}
}

func TestPublicClient_Ping_OK_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" {
			t.Fatalf("expected /ping, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true,"message":"pong"}`))
	}))
	defer srv.Close()

	c, err := NewPublicClient(srv.URL) // http ok for tests
	if err != nil {
		t.Fatalf("NewPublicClient err: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := c.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping err: %v", err)
	}
	if out == nil || !out.OK || out.Message != "pong" {
		t.Fatalf("unexpected ping response: %+v", out)
	}
}

func TestPublicClient_Ping_Non2xx_ReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`rate limited`))
	}))
	defer srv.Close()

	c, err := NewPublicClient(srv.URL)
	if err != nil {
		t.Fatalf("NewPublicClient err: %v", err)
	}

	_, err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	// must be a typed Status error
	var st *types.Status
	if !errors.As(err, &st) {
		t.Fatalf("expected errors.As to find *types.Status, got %T %v", err, err)
	}
	if st.StatusCode != 429 {
		t.Fatalf("expected 429, got %d", st.StatusCode)
	}
	if st.Path != "/ping" {
		t.Fatalf("expected path /ping, got %s", st.Path)
	}

	// top level kind should be KindStatus
	var top *types.Error
	if !errors.As(err, &top) {
		t.Fatalf("expected errors.As to find *types.Error, got %T %v", err, err)
	}
	if top.Kind() != types.KindStatus {
		t.Fatalf("expected kind %s, got %s", types.KindStatus, top.Kind())
	}
}
