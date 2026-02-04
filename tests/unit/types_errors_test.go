package unit_test

import (
	"errors"
	// "strings"
	"testing"

	"github.com/star-gazer111/poly-go-clob-client/types"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         []byte
		wantNil      bool
		wantSentinel error
		wantCode     string
		wantMsg      string
	}{
		{
			name:    "success 200 returns nil",
			status:  200,
			body:    []byte(`{"result": "ok"}`),
			wantNil: true,
		},
		{
			name:    "success 201 returns nil",
			status:  201,
			body:    []byte(`{"created": true}`),
			wantNil: true,
		},
		{
			name:         "rate limit 429 with message and code",
			status:       429,
			body:         []byte(`{"message": "too many requests", "code": "RATE_LIMIT"}`),
			wantSentinel: types.ErrRateLimited,
			wantCode:     "RATE_LIMIT",
			wantMsg:      "too many requests",
		},
		{
			name:         "unauthorized 401 with error field",
			status:       401,
			body:         []byte(`{"error": "invalid credentials", "code": "AUTH_FAILED"}`),
			wantSentinel: types.ErrUnauthorized,
			wantCode:     "AUTH_FAILED",
			wantMsg:      "invalid credentials",
		},
		{
			name:         "forbidden 403 maps to unauthorized",
			status:       403,
			body:         []byte(`{"message": "access denied"}`),
			wantSentinel: types.ErrUnauthorized,
			wantCode:     "",
			wantMsg:      "access denied",
		},
		{
			name:         "bad request 400",
			status:       400,
			body:         []byte(`{"message": "invalid input", "code": "VALIDATION_ERROR"}`),
			wantSentinel: types.ErrBadRequest,
			wantCode:     "VALIDATION_ERROR",
			wantMsg:      "invalid input",
		},
		{
			name:         "server error 500",
			status:       500,
			body:         []byte(`{"message": "internal server error"}`),
			wantSentinel: types.ErrServer,
			wantCode:     "",
			wantMsg:      "internal server error",
		},
		{
			name:         "server error 503 service unavailable",
			status:       503,
			body:         []byte(`{"error": "service unavailable"}`),
			wantSentinel: types.ErrServer,
			wantCode:     "",
			wantMsg:      "service unavailable",
		},
		{
			name:         "malformed JSON falls back to bad request",
			status:       418,
			body:         []byte(`not valid json`),
			wantSentinel: types.ErrBadRequest,
			wantCode:     "",
			wantMsg:      "",
		},
		{
			name:         "empty body",
			status:       500,
			body:         []byte(``),
			wantSentinel: types.ErrServer,
			wantCode:     "",
			wantMsg:      "",
		},
	}

	for _,tt := range tests{
		t.Run(tt.name, func(t *testing.T){
			err := types.Classify(tt.status, tt.body)

			if tt.wantNil{
				if err != nil{
					t.Errorf("expected nil error for success status, got %v", err)
				}
				return
			}

			if err == nil{
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, tt.wantSentinel){
				t.Errorf("expected sentinel %v, got %v", tt.wantSentinel, err)
			}

			apiErr, ok := err.(*types.APIError)
			if !ok{
				t.Fatalf("error is not *types.APIError, got %v", apiErr)
			}

			if apiErr.Status != tt.status{
				t.Errorf("expected status : %v, got status : %v", tt.status, apiErr.Status)
			}

			if apiErr.Code != tt.wantCode{
				t.Errorf("expected status : %v, got status : %v", tt.wantCode, apiErr.Code)
			}

			if apiErr.Message != tt.wantMsg{
				t.Errorf("expected status : %v, got status : %v", tt.wantMsg, apiErr.Message)
			}

			if string(apiErr.RawBody) != string(tt.body){
				t.Errorf("RawBody is not preserved correctly")
			}

		})
	}
}
