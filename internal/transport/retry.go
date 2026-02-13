package transport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/types"
)

type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  150 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	}
}

// Conservative but correct idempotency set.
// PUT/DELETE are idempotent per HTTP semantics.
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func shouldRetryStatus(status int) bool {
	return status == 429 || (status >= 500 && status <= 599)
}

// ensureReplayableBody ensures req.GetBody is set (so we can reset req.Body before each attempt).
// This avoids the classic bug: retries send an empty body after the first attempt.
func ensureReplayableBody(req *http.Request) error {
	if req == nil || req.Body == nil {
		return nil
	}
	if req.GetBody != nil {
		return nil
	}

	// Read the original body once into memory.
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	_ = req.Body.Close()

	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(b)), nil
	}

	// Reset body for first attempt too.
	rc, err := req.GetBody()
	if err != nil {
		return err
	}
	req.Body = rc
	req.ContentLength = int64(len(b))
	return nil
}

func resetBody(req *http.Request) error {
	if req == nil || req.GetBody == nil {
		return nil
	}
	rc, err := req.GetBody()
	if err != nil {
		return err
	}
	req.Body = rc
	return nil
}

func bodyPreview(b []byte) string {
	if len(b) > types.MaxRawBodyBytes {
		b = b[:types.MaxRawBodyBytes]
	}
	return string(b)
}

// doJSONWithRetry performs the request via Transport.Do and returns the raw response body.
// On non-2xx it returns a typed error compatible with:
//   errors.As(err, *types.Status)
//   errors.As(err, *types.Error) with KindStatus
func doJSONWithRetry(ctx context.Context, t *Transport, req *http.Request) ([]byte, error) {
	p := t.policy.Retry
	attempts := 0

	// Make sure the request body can be replayed.
	if err := ensureReplayableBody(req); err != nil {
		return nil, types.WithSource(types.KindInternal, err)
	}

	for {
		// Reset body before each attempt so retries send the same payload.
		if err := resetBody(req); err != nil {
			return nil, types.WithSource(types.KindInternal, err)
		}

		resp, err := t.Do(ctx, req)
		if err != nil {
			// Only retry network-ish failures for idempotent methods.
			if !isIdempotent(req.Method) || attempts >= p.MaxRetries {
				return nil, types.WithSource(types.KindInternal, err)
			}
			attempts++
			sleepBackoff(ctx, p, attempts)
			continue
		}

		// IMPORTANT: close per attempt; do NOT defer in loop.
		b, rerr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if rerr != nil {
			return nil, types.WithSource(types.KindInternal, rerr)
		}

		// Your Transport.Do wraps the body with MaxBodyBytes+1. Detect overflow here.
		if t.MaxBodyBytes() > 0 && int64(len(b)) > t.MaxBodyBytes() {
			return nil, types.WithSource(types.KindInternal, types.ErrBodyTooLarge)
		}

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return b, nil
		}

		// Build a structured Status error (KindStatus) like Rust.
		statusErr := types.StatusErr(resp.StatusCode, req.Method, req.URL.Path, bodyPreview(b))

		// Retry only on idempotent + transient status.
		if !isIdempotent(req.Method) || attempts >= p.MaxRetries || !shouldRetryStatus(resp.StatusCode) {
			return nil, statusErr
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
