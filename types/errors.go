package types

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)

// kind matches the Rust enum Kind
type Kind string

var ErrBodyTooLarge = errors.New("response body too large")

const (
	KindStatus          Kind = "status"
	KindValidation      Kind = "validation"
	KindSynchronization Kind = "synchronization"
	KindInternal        Kind = "internal"
	KindWebSocket       Kind = "websocket"
	KindGeoblock        Kind = "geoblock"
)

// error is the top-level error wrapper
type Error struct {
	kind  Kind
	src   error
	stack []byte // lightweight equivalent of backtrace
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.src != nil {
		return fmt.Sprintf("%s: %s", e.kind, e.src.Error())
	}
	return string(e.kind)
}

func (e *Error) Unwrap() error { return e.src }

func (e *Error) Kind() Kind { return e.kind }

// stack returns a captured stack trace (if enabled)
func (e *Error) Stack() []byte { return e.stack }

// WithSource mirrors Rust's Error::with_source
func WithSource(kind Kind, src error) *Error {
	if src == nil {
		return &Error{kind: kind}
	}
	return &Error{kind: kind, src: src, stack: debug.Stack()}
}

// -------- Structured inner error types --------

type Status struct {
	StatusCode int
	Method     string
	Path       string
	Message    string
}

func (s *Status) Error() string {
	method := s.Method
	if method == "" {
		method = "UNKNOWN"
	}
	msg := strings.TrimSpace(s.Message)
	if msg == "" {
		msg = "unknown error"
	}
	return fmt.Sprintf("error(%d) making %s call to %s with %s", s.StatusCode, method, s.Path, msg)
}

type Validation struct {
	Reason string
}

func (v *Validation) Error() string {
	return fmt.Sprintf("invalid: %s", strings.TrimSpace(v.Reason))
}

type Synchronization struct{}

func (s *Synchronization) Error() string {
	return "synchronization error: multiple goroutines are attempting to log in or log out"
}

// in rust chain_id is ChainId but in go i am keeping it uint64 for now - confirm later todo
type MissingContractConfig struct {
	ChainID uint64
	NegRisk bool
}

func (m *MissingContractConfig) Error() string {
	return fmt.Sprintf("missing contract config for chain id %d with neg_risk = %t", m.ChainID, m.NegRisk)
}

type Geoblock struct {
	IP      string
	Country string
	Region  string
}

func (g *Geoblock) Error() string {
	return fmt.Sprintf("access blocked from country: %s, region: %s, ip: %s", g.Country, g.Region, g.IP)
}

func ValidationErr(reason string) *Error {
	return WithSource(KindValidation, &Validation{Reason: reason})
}

func StatusErr(statusCode int, method string, path string, message string) *Error {
	return WithSource(KindStatus, &Status{
		StatusCode: statusCode,
		Method:     method,
		Path:       path,
		Message:    message,
	})
}

func SyncErr() *Error {
	return WithSource(KindSynchronization, &Synchronization{})
}

func MissingContractConfigErr(chainID uint64, negRisk bool) *Error {
	return WithSource(KindInternal, &MissingContractConfig{
		ChainID: chainID,
		NegRisk: negRisk,
	})
}

func GeoblockErr(ip, country, region string) *Error {
	return WithSource(KindGeoblock, &Geoblock{
		IP:      ip,
		Country: country,
		Region:  region,
	})
}

func AsStatus(err error) (*Status, bool) {
	var s *Status
	if errors.As(err, &s) {
		return s, true
	}
	return nil, false
}

func AsValidation(err error) (*Validation, bool) {
	var v *Validation
	if errors.As(err, &v) {
		return v, true
	}
	return nil, false
}

// -------- HTTP classification --------

// MaxRawBodyBytes caps what we include in messages to avoid giant bodies/PII
const MaxRawBodyBytes = 4 << 10 // 4KiB

// HTTPErrorBody is a best effort parse of common error fields
type HTTPErrorBody struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Code    any    `json:"code"`
}

// ClassifyHTTP turns a non-2xx HTTP response into a structured Status error wrapped with KindStatus
// this mirrors the rust’s "Status kind" while still letting us inspect status/method/path
func ClassifyHTTP(statusCode int, method string, path string, bodyPreview string) error {
	msg := strings.TrimSpace(bodyPreview)
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	return StatusErr(statusCode, method, path, msg)
}
