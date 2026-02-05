package redaction

import (
	"net/http"
	"strings"
)

// DefaultRedaction is shown when a value is fully hidden
const DefaultRedaction = "***"

// Redact returns a redacted version of s
// It preserves a small prefix/suffix to help debugging without leaking secrets
//
// Some simple examples to understand:
//   "abcd" -> "***"
//   "abcdefg" -> "abc***efg"
func Redact(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Keeping the behavior conservative for short secrets 
	if len(s) <= 8 {
		return DefaultRedaction
	}
	// Keeping the first 3 and last 3 chars
	return s[:3] + DefaultRedaction + s[len(s)-3:]
}

// RedactHeaders returns a copy of headers with sensitive values redacted & it never mutates the input header map
//
// It redacts by header name and also by token patterns for unknown keys
func RedactHeaders(h http.Header) http.Header {
	if h == nil {
		return nil
	}

	out := make(http.Header, len(h))
	for k, vv := range h {
		ck := http.CanonicalHeaderKey(k)

		// always copy the slice to avoid aliasing
		copied := make([]string, 0, len(vv))

		// redact based on known sensitive header names
		if isSensitiveHeaderName(ck) {
			for range vv {
				copied = append(copied, DefaultRedaction)
			}
			out[ck] = copied
			continue
		}

		// for other headers redact if they look like secrets (e.g., Bearer tokens)
		for _, v := range vv {
			copied = append(copied, redactHeaderValueHeuristic(v))
		}
		out[ck] = copied
	}

	return out
}

// isSensitiveHeaderName returns true for headers that commonly carry secrets
func isSensitiveHeaderName(canonicalKey string) bool {
	switch canonicalKey {
	case "Authorization",
		"Proxy-Authorization",
		"Cookie",
		"Set-Cookie",
		"X-Api-Key",
		"Api-Key",
		"X-Api-Token",
		"X-Auth-Token",
		"X-Access-Token",
		// Polymarket specific-ish (can expand later)
		"Poly-Api-Key",
		"Poly-Api-Secret",
		"Poly-Api-Passphrase":
		return true
	default:
		// Heuristic: any header name containing these substrings is treated as sensitive
		lk := strings.ToLower(canonicalKey)
		if strings.Contains(lk, "secret") ||
			strings.Contains(lk, "token") ||
			strings.Contains(lk, "pass") ||
			strings.Contains(lk, "key") ||
			strings.Contains(lk, "session") {
			return true
		}
		return false
	}
}

// redactHeaderValueHeuristic tries to avoid leaking secrets for unknown keys
// - "Bearer <token>" -> "Bearer ***"
// - "<very long string>" -> prefix/suffix redacted
func redactHeaderValueHeuristic(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(v), "bearer ") {
		return "Bearer " + DefaultRedaction
	}

	// if it looks long/secretish, redact with partial reveal
	if len(v) >= 16 {
		return Redact(v)
	}

	return v
}
