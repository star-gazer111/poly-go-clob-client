package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	// Retry only safe/idempotent methods by default.
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  150 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	}
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func shouldRetry(method string, status int, err error) bool {
	if err != nil {
		// network-ish errors: retry only if idempotent
		return true
	}
	// Retry 429/5xx only for idempotent methods in v0.1
	if status == 429 || (status >= 500 && status <= 599) {
		return true
	}
	return false
}

func doJSONWithRetry(ctx context.Context, t *Transport, req *http.Request) ([]byte, error) {
	p := t.policy.Retry
	attempts := 0

	for {
		resp, err := t.Do(ctx, req)
		if err != nil {
			if !isIdempotent(req.Method) || attempts >= p.MaxRetries {
				return nil, err
			}
			attempts++
			sleepBackoff(ctx, p, attempts)
			continue
		}

		// Body is already wrapped with proper cancellation and limit.
		// We just need to read it fully.
		defer resp.Body.Close()

		b, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return nil, rerr
		}

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return b, nil
		}

		if !isIdempotent(req.Method) || attempts >= p.MaxRetries || !shouldRetry(req.Method, resp.StatusCode, nil) {
			return nil, errors.New(string(b))
		}

		attempts++
		sleepBackoff(ctx, p, attempts)
	}
}

func sleepBackoff(ctx context.Context, p RetryPolicy, attempt int) {
	delay := p.BaseDelay * time.Duration(1<<uint(attempt-1))
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
