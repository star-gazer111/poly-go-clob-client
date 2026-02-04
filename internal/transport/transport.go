package transport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// these are the rules the transport layer must follow
type Policy struct {
	Timeout      time.Duration
	RateLimiter  *rate.Limiter
	MaxBodyBytes int64
	Retry        RetryPolicy
	UserAgent    string
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

// basically checks whether all policies are followed or not
func (t *Transport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if t.policy.RateLimiter != nil {
		if err := t.policy.RateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	// We cannot use defer cancel() here because it would cancel the context
	// immediately after headers are received, breaking body reads.
	// Instead, we wrap the body to cancel on Close().
	ctx, cancel := context.WithTimeout(ctx, t.policy.Timeout)

	req = req.WithContext(ctx)
	if t.policy.UserAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.policy.UserAgent)
	}

	resp, err := t.hc.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	resp.Body = &bodyWrapper{
		ReadCloser: resp.Body,
		cancel:     cancel,
		reader:     io.LimitReader(resp.Body, t.policy.MaxBodyBytes+1),
	}

	return resp, nil
}

type bodyWrapper struct {
	io.ReadCloser
	cancel context.CancelFunc
	reader io.Reader
}

func (w *bodyWrapper) Read(p []byte) (int, error) {
	return w.reader.Read(p)
}

func (w *bodyWrapper) Close() error {
	defer w.cancel()
	return w.ReadCloser.Close()
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

func (t *Transport) MaxBodyBytes() int64 {
	return t.policy.MaxBodyBytes
}
