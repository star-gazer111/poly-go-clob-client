package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// define Sentinel Errors
var (
	ErrRateLimited = errors.New("rate limited")
	ErrUnauthorized = errors.New("unauthorized")
	ErrBadRequest = errors.New("bad request")
	ErrServer = errors.New("server error")
	ErrUnknown = errors.New("unknown error")
)

type APIError struct{
	Status int 		// http status code
	Code string 	// error code from Polymarket
	Message string  // human-readable message
	RawBody []byte  // Full response body
	wrapped error	// the sentinel error this wraps
}

func (e *APIError) Error() string{
	bodyPreview := string(e.RawBody)
	if len(bodyPreview) > 200{
		bodyPreview = bodyPreview[:200] + "...(truncated)"
	}

	return fmt.Sprintf("API error: status=%d, code=%s, message=%s, body=%s",
		e.Status, e.Code, e.Message, bodyPreview)
}

// this makes `errors.Is()` work for APIError struct
func (e *APIError) Unwrap() error{
	return e.wrapped
}

func Classify(status int, body []byte) error{
	if status >= 200 && status < 300{
		return nil
	}

	var parsed map[string]interface{}
	_ = json.Unmarshal(body, &parsed)

	message := ""
	code := ""

	if msg,ok := parsed["message"].(string); ok{
		message = msg
	}

	if message == ""{
		if err,ok := parsed["error"].(string); ok{
			message = err
		}
	}

	if c,ok := parsed["code"].(string); ok{
		code = c
	}

	var sentinel error
	switch{
	case status == 429:
		sentinel = ErrRateLimited
	case status == 401 || status == 403:
		sentinel = ErrUnauthorized
	case status >= 400 && status < 500:
		sentinel = ErrBadRequest
	case status >= 500:
		sentinel = ErrServer
	default:
		sentinel = ErrUnknown
	}

	return &APIError{
		Status : status,
		Code : code,
		Message: message,
		RawBody: body,
		wrapped: sentinel,
	}
}