package types

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestStatusErrKindAndAs(t *testing.T) {
	err := StatusErr(429, http.MethodGet, "/markets", "rate limited")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// top-level error
	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error, got %T", err)
	}
	if top.Kind() != KindStatus {
		t.Fatalf("expected kind %s, got %s", KindStatus, top.Kind())
	}

	// inner Status error
	var st *Status
	if !errors.As(err, &st) {
		t.Fatalf("expected *types.Status via errors.As")
	}
	if st.StatusCode != 429 {
		t.Fatalf("expected status 429, got %d", st.StatusCode)
	}
	if st.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %s", st.Method)
	}
	if st.Path != "/markets" {
		t.Fatalf("expected path /markets, got %s", st.Path)
	}

	// display format sanity
	msg := err.Error()
	if !strings.Contains(msg, "429") || !strings.Contains(msg, "GET") {
		t.Fatalf("unexpected Error() output: %s", msg)
	}
}

func TestValidationErr(t *testing.T) {
	err := ValidationErr("invalid price precision")

	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error")
	}
	if top.Kind() != KindValidation {
		t.Fatalf("expected kind %s, got %s", KindValidation, top.Kind())
	}

	var v *Validation
	if !errors.As(err, &v) {
		t.Fatalf("expected *types.Validation")
	}
	if v.Reason != "invalid price precision" {
		t.Fatalf("unexpected reason: %s", v.Reason)
	}
}

func TestSynchronizationErr(t *testing.T) {
	err := SyncErr()

	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error")
	}
	if top.Kind() != KindSynchronization {
		t.Fatalf("expected kind %s, got %s", KindSynchronization, top.Kind())
	}

	var s *Synchronization
	if !errors.As(err, &s) {
		t.Fatalf("expected *types.Synchronization")
	}

	if !strings.Contains(err.Error(), "synchronization") {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}

func TestGeoblockErr(t *testing.T) {
	err := GeoblockErr("192.168.1.1", "US", "NY")

	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error")
	}
	if top.Kind() != KindGeoblock {
		t.Fatalf("expected kind %s, got %s", KindGeoblock, top.Kind())
	}

	var g *Geoblock
	if !errors.As(err, &g) {
		t.Fatalf("expected *types.Geoblock")
	}
	if g.Country != "US" || g.Region != "NY" {
		t.Fatalf("unexpected geoblock fields: %+v", g)
	}

	if !strings.Contains(err.Error(), "US") {
		t.Fatalf("expected country in error string")
	}
}

func TestMissingContractConfigErr(t *testing.T) {
	err := MissingContractConfigErr(137, true)

	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error")
	}
	if top.Kind() != KindInternal {
		t.Fatalf("expected kind %s, got %s", KindInternal, top.Kind())
	}

	var m *MissingContractConfig
	if !errors.As(err, &m) {
		t.Fatalf("expected *types.MissingContractConfig")
	}
	if m.ChainID != 137 || !m.NegRisk {
		t.Fatalf("unexpected fields: %+v", m)
	}
}

func TestWithSourceCapturesStack(t *testing.T) {
	src := errors.New("inner failure")
	err := WithSource(KindInternal, src)

	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error")
	}
	if top.Stack() == nil || len(top.Stack()) == 0 {
		t.Fatalf("expected stack trace to be captured")
	}
}

func TestClassifyHTTPProducesStatusErr(t *testing.T) {
	err := ClassifyHTTP(403, http.MethodPost, "/orders", "unauthorized")

	var top *Error
	if !errors.As(err, &top) {
		t.Fatalf("expected *types.Error")
	}
	if top.Kind() != KindStatus {
		t.Fatalf("expected kind %s, got %s", KindStatus, top.Kind())
	}

	var st *Status
	if !errors.As(err, &st) {
		t.Fatalf("expected *types.Status")
	}
	if st.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", st.StatusCode)
	}
}
