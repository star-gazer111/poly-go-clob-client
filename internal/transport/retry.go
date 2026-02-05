package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
	"log"
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

defer func() {
	if err := resp.Body.Close(); err != nil {
		// best-effort; close errors are rarely actionable but can be logged
		log.Printf("resp body close: %v", err)
	}
}()


		// Read with size cap
		b, rerr := readAllCapped(resp.Body, t.policy.MaxBodyBytes)
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
	if attempt <= 0 {
		attempt = 1
	}

	delay := p.BaseDelay
	if delay <= 0 {
		delay = 50 * time.Millisecond // or whatever default you want
	}
	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	// Exponential: delay *= 2^(attempt-1), but clamp to MaxDelay safely.
	for i := 1; i < attempt; i++ {
		if p.MaxDelay > 0 {
			// If doubling would exceed MaxDelay, clamp and stop.
			if delay >= p.MaxDelay/2 {
				delay = p.MaxDelay
				break
			}
		}
		// Doubling duration is safe here because we clamp above.
		delay *= 2
	}

	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	t := time.NewTimer(delay)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return
	case <-t.C:
		return
	}
}

func readAllCapped(r io.Reader, max int64) ([]byte, error) {
	// minimal capped reader
	lr := &io.LimitedReader{R: r, N: max + 1}
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, errors.New("response body too large")
	}
	return b, nil
}
