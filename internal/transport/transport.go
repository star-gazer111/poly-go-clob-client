package transport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Policy struct {
	Timeout       time.Duration
	RateLimiter   *rate.Limiter
	MaxBodyBytes  int64
	Retry         RetryPolicy
	UserAgent     string
}

func DefaultPolicy() Policy {
	return Policy{
		Timeout:      12 * time.Second,
		RateLimiter:  rate.NewLimiter(rate.Limit(8), 16), // conservative default
		MaxBodyBytes: 2 << 20,                            // 2 MiB
		Retry:        DefaultRetryPolicy(),
		UserAgent:    "poly-go-clob-client/0.1",
	}
}

type Transport struct {
	hc     *http.Client
	policy Policy
}

func NewTransport(hc *http.Client, p Policy) *Transport {
	// Ensure hc has no zero timeouts (we enforce via context timeout anyway)
	return &Transport{hc: hc, policy: p}
}

func (t *Transport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if t.policy.RateLimiter != nil {
		if err := t.policy.RateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, t.policy.Timeout)
	defer cancel()

	req = req.WithContext(ctx)
	if t.policy.UserAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.policy.UserAgent)
	}

	return t.hc.Do(req)
}

func (t *Transport) DoJSON(ctx context.Context, method, url string, headers map[string]string, body []byte) ([]byte, error) {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return doJSONWithRetry(ctx, t, req)
}
