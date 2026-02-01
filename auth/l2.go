package auth

import "net/http"

// ApplyL2Headers applies Polymarket API key headers to a request.
// Exact header names/format should match official clients.
func ApplyL2Headers(req *http.Request, creds APICreds) {
	// TODO: req.Header.Set("POLY_API_KEY", creds.Key) etc.
	// Ensure that we NEVER log these values.
}
