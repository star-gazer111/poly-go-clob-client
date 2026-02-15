package clob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/star-gazer111/poly-go-clob-client/types"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *PublicClient) {
	srv := httptest.NewServer(handler)
	c, err := NewPublicClient(srv.URL)
	if err != nil {
		t.Fatalf("NewPublicClient err: %v", err)
	}
	return srv, c
}

func TestMidpoint_Happy(t *testing.T) {
	expectedMid := decimal.NewFromFloat(0.5)
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/midpoint" {
			t.Errorf("Expected /midpoint, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "123" {
			t.Errorf("Expected token_id=123, got %s", r.URL.Query().Get("token_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		resp := types.MidpointResponse{Mid: expectedMid}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	resp, err := c.Midpoint(context.Background(), &types.MidpointRequest{TokenId: "123"})
	if err != nil {
		t.Fatalf("Midpoint err: %v", err)
	}
	if !resp.Mid.Equal(expectedMid) {
		t.Errorf("Expected %v, got %v", expectedMid, resp.Mid)
	}
}

func TestMidpoint_Errors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		kind       types.Kind
	}{
		{"429", 429, types.KindStatus},
		{"500", 500, types.KindStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})
			defer srv.Close()

			_, err := c.Midpoint(context.Background(), &types.MidpointRequest{TokenId: "123"})
			if err == nil {
				t.Fatal("Expected error")
			}
			var te *types.Error
			if errors.As(err, &te) {
				if te.Kind() != tt.kind {
					t.Errorf("Expected kind %v, got %v", tt.kind, te.Kind())
				}
			}
		})
	}
}

func TestMidpoints_Happy(t *testing.T) {
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/midpoints" {
			t.Errorf("Expected /midpoints, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		// Correct response strictly is map directly
		m := map[string]decimal.Decimal{
			"123": decimal.NewFromFloat(0.5),
		}
		json.NewEncoder(w).Encode(m)
	})
	defer srv.Close()

	req := []types.MidpointRequest{{TokenId: "123"}}
	resp, err := c.Midpoints(context.Background(), req)
	if err != nil {
		t.Fatalf("Midpoints err: %v", err)
	}
	if val, ok := resp["123"]; !ok || !val.Equal(decimal.NewFromFloat(0.5)) {
		t.Errorf("Expected 123:0.5, got %v", resp)
	}
}

func TestMidpoints_Errors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"429", 429},
		{"500", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})
			defer srv.Close()
			_, err := c.Midpoints(context.Background(), []types.MidpointRequest{{TokenId: "123"}})
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestGetPrice_Happy(t *testing.T) {
	expectedPrice := decimal.NewFromFloat(0.4)
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price" {
			t.Errorf("Expected /price, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "123" {
			t.Errorf("Expected token_id=123")
		}
		if r.URL.Query().Get("side") != "BUY" {
			t.Errorf("Expected side=BUY")
		}
		json.NewEncoder(w).Encode(types.PriceResponse{Price: expectedPrice})
	})
	defer srv.Close()

	resp, err := c.GetPrice(context.Background(), types.PriceRequest{TokenId: "123", Side: "BUY"})
	if err != nil {
		t.Fatalf("GetPrice err: %v", err)
	}
	if !resp.Price.Equal(expectedPrice) {
		t.Errorf("Expected %v, got %v", expectedPrice, resp.Price)
	}
}

func TestGetPrice_Errors(t *testing.T) {
	tests := []struct{ code int }{{429}, {500}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.code) })
			defer srv.Close()
			_, err := c.GetPrice(context.Background(), types.PriceRequest{TokenId: "123", Side: "BUY"})
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestGetPrices_Happy(t *testing.T) {
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prices" {
			t.Errorf("Expected /prices")
		}
		m := map[string]map[string]decimal.Decimal{
			"123": {"BUY": decimal.NewFromFloat(0.4)},
		}
		json.NewEncoder(w).Encode(m)
	})
	defer srv.Close()

	req := []types.PriceRequest{{TokenId: "123", Side: "BUY"}}
	resp, err := c.GetPrices(context.Background(), req)
	if err != nil {
		t.Fatalf("GetPrices err: %v", err)
	}
	if val := resp["123"]["BUY"]; !val.Equal(decimal.NewFromFloat(0.4)) {
		t.Errorf("Expected 0.4, got %v", val)
	}
}

func TestGetPrices_Errors(t *testing.T) {
	tests := []struct{ code int }{{429}, {500}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.code) })
			defer srv.Close()
			_, err := c.GetPrices(context.Background(), nil)
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestGetSpread_Happy(t *testing.T) {
	expectedSpread := decimal.NewFromFloat(0.01)
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/spread" {
			t.Errorf("Expected /spread")
		}
		if r.URL.Query().Get("token_id") != "123" {
			t.Errorf("Expected token_id=123")
		}
		json.NewEncoder(w).Encode(types.SpreadResponse{Spread: expectedSpread})
	})
	defer srv.Close()

	resp, err := c.GetSpread(context.Background(), types.SpreadRequest{TokenId: "123"})
	if err != nil {
		t.Fatalf("GetSpread err: %v", err)
	}
	if !resp.Spread.Equal(expectedSpread) {
		t.Errorf("Expected %v, got %v", expectedSpread, resp.Spread)
	}
}

func TestGetSpread_Errors(t *testing.T) {
	tests := []struct{ code int }{{429}, {500}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.code) })
			defer srv.Close()
			_, err := c.GetSpread(context.Background(), types.SpreadRequest{TokenId: "123"})
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestGetSpreads_Happy(t *testing.T) {
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/spreads" {
			t.Errorf("Expected /spreads")
		}
		m := map[string]decimal.Decimal{
			"123": decimal.NewFromFloat(0.01),
		}
		json.NewEncoder(w).Encode(m)
	})
	defer srv.Close()

	resp, err := c.GetSpreads(context.Background(), []types.SpreadRequest{{TokenId: "123"}})
	if err != nil {
		t.Fatalf("GetSpreads err: %v", err)
	}
	if val := resp["123"]; !val.Equal(decimal.NewFromFloat(0.01)) {
		t.Errorf("Expected 0.01, got %v", val)
	}
}

func TestGetSpreads_Errors(t *testing.T) {
	tests := []struct{ code int }{{429}, {500}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.code) })
			defer srv.Close()
			_, err := c.GetSpreads(context.Background(), []types.SpreadRequest{{TokenId: "123"}})
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestGetPricesHistory_Happy(t *testing.T) {
	srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prices-history" {
			t.Errorf("Expected /prices-history")
		}
		if r.URL.Query().Get("market") != "123" {
			t.Errorf("Expected market=123")
		}
		if r.URL.Query().Get("interval") != "1m" {
			t.Errorf("Expected interval=1m")
		}

		item := types.PriceHistoryItem{
			Time:  time.Now().Unix(),
			Price: decimal.NewFromFloat(0.5),
		}
		resp := types.PricesHistoryResponse{
			History: []types.PriceHistoryItem{item},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	req := types.PricesHistoryRequest{
		Market:   "123",
		Interval: types.Interval1m,
	}
	resp, err := c.GetPricesHistory(context.Background(), req)
	if err != nil {
		t.Fatalf("GetPricesHistory err: %v", err)
	}
	if len(resp.History) != 1 {
		t.Errorf("Expected 1 history item, got %d", len(resp.History))
	}
}

func TestGetPricesHistory_Validation(t *testing.T) {
	// Not an HTTP test, but validation logic test
	c, _ := NewPublicClient("http://example.com")

	// Default: Fail (must provide interval or range)
	_, err := c.GetPricesHistory(context.Background(), types.PricesHistoryRequest{Market: "123"})
	if err == nil {
		t.Error("Expected error for missing Interval/Range")
	}

	// Double: Fail (cannot provide both)
	start := int64(100)
	_, err = c.GetPricesHistory(context.Background(), types.PricesHistoryRequest{
		Market:   "123",
		Interval: types.Interval1m,
		StartTs:  &start,
	})
	if err == nil {
		t.Error("Expected error for both Interval and Range")
	}
}

func TestGetPricesHistory_Errors(t *testing.T) {
	tests := []struct{ code int }{{429}, {500}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			srv, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.code) })
			defer srv.Close()
			_, err := c.GetPricesHistory(context.Background(), types.PricesHistoryRequest{Market: "123", Interval: types.Interval1m})
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}
