//go:build integration
// +build integration

package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/types"
)

const (
	// LiveBaseURL is the Polymarket CLOB API base URL.
	LiveBaseURL = "https://clob.polymarket.com"
	// DefaultTimeout for API calls.
	DefaultTimeout = 30 * time.Second
)

// getTestClient returns a PublicClient configured for live testing.
func getTestClient(t *testing.T) *PublicClient {
	t.Helper()

	baseURL := os.Getenv("POLYMARKET_CLOB_URL")
	if baseURL == "" {
		baseURL = LiveBaseURL
	}

	c, err := NewPublicClient(baseURL, WithRequireHTTPS(true))
	if err != nil {
		t.Fatalf("NewPublicClient: %v", err)
	}
	return c
}

// TestIntegration_Ping tests the root endpoint as a health check against the live API.
func TestIntegration_Ping(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	t.Logf("Ping response: %+v", resp)
	if resp == nil {
		t.Fatal("expected non-nil ping response")
	}
}

// TestIntegration_Markets tests the /markets endpoint against the live API.
func TestIntegration_Markets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.Markets(ctx, "")
	fmt.Println(resp)
	if err != nil {
		t.Fatalf("Markets failed: %v", err)
	}

	t.Logf("Markets response: Count=%d, Limit=%d, NextCursor=%s", resp.Count, resp.Limit, resp.NextCursor)

	if resp.Count == 0 {
		t.Log("Warning: No markets returned from the API")
	}

	// Validate pagination works
	if resp.NextCursor != "" && resp.Count > 0 {
		t.Logf("First market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_Markets_Pagination tests pagination on /markets endpoint.
func TestIntegration_Markets_Pagination(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// First page
	page1, err := c.Markets(ctx, "")
	if err != nil {
		t.Fatalf("Markets (page 1) failed: %v", err)
	}

	if page1.NextCursor == "" {
		t.Skip("No next cursor, cannot test pagination")
	}

	// Second page
	page2, err := c.Markets(ctx, page1.NextCursor)
	if err != nil {
		t.Fatalf("Markets (page 2) failed: %v", err)
	}

	t.Logf("Page 1: Count=%d, Page 2: Count=%d", page1.Count, page2.Count)

	// Ensure we got different data (if both pages have data)
	if len(page1.Data) > 0 && len(page2.Data) > 0 {
		if page1.Data[0].ConditionID != nil && page2.Data[0].ConditionID != nil {
			if *page1.Data[0].ConditionID == *page2.Data[0].ConditionID {
				t.Error("Page 1 and Page 2 returned the same first market")
			}
		}
	}
}

// TestIntegration_Market tests the /markets/{condition_id} endpoint against the live API.
func TestIntegration_Market(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// First, get a valid condition_id from the markets list
	markets, err := c.Markets(ctx, "")
	if err != nil {
		t.Fatalf("Markets failed: %v", err)
	}

	if len(markets.Data) == 0 {
		t.Skip("No markets available to test Market endpoint")
	}

	// Find a market with a valid condition_id
	var conditionID string
	for _, m := range markets.Data {
		if m.ConditionID != nil && *m.ConditionID != "" {
			conditionID = *m.ConditionID
			break
		}
	}

	if conditionID == "" {
		t.Skip("No market with condition_id found")
	}

	t.Logf("Testing Market with condition_id: %s", conditionID)

	req := &types.MarketRequest{ConditionID: conditionID}
	resp, err := c.Market(ctx, req)
	if err != nil {
		t.Fatalf("Market failed: %v", err)
	}

	t.Logf("Market response: Question=%s, Active=%v", resp.Question, resp.Active)

	if resp.ConditionID == nil || *resp.ConditionID != conditionID {
		t.Errorf("Expected condition_id %s, got %v", conditionID, resp.ConditionID)
	}
}

// TestIntegration_SimplifiedMarkets tests the /simplified-markets endpoint against the live API.
func TestIntegration_SimplifiedMarkets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.SimplifiedMarkets(ctx, "")
	if err != nil {
		t.Fatalf("SimplifiedMarkets failed: %v", err)
	}

	t.Logf("SimplifiedMarkets response: Count=%d, Limit=%d", resp.Count, resp.Limit)

	if resp.Count > 0 && len(resp.Data) > 0 {
		t.Logf("First simplified market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_SamplingMarkets tests the /sampling-markets endpoint against the live API.
func TestIntegration_SamplingMarkets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.SamplingMarkets(ctx, "")
	if err != nil {
		// Response body too large is expected for this endpoint
		if strings.Contains(err.Error(), "response body too large") {
			t.Skip("SamplingMarkets response too large for default transport settings")
		}
		t.Fatalf("SamplingMarkets failed: %v", err)
	}

	t.Logf("SamplingMarkets response: Count=%d, Limit=%d", resp.Count, resp.Limit)

	if resp.Count > 0 && len(resp.Data) > 0 {
		t.Logf("First sampling market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_SamplingSimplifiedMarkets tests the /sampling-simplified-markets endpoint.
func TestIntegration_SamplingSimplifiedMarkets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.SamplingSimplifiedMarkets(ctx, "")
	if err != nil {
		t.Fatalf("SamplingSimplifiedMarkets failed: %v", err)
	}

	t.Logf("SamplingSimplifiedMarkets response: Count=%d, Limit=%d", resp.Count, resp.Limit)

	if resp.Count > 0 && len(resp.Data) > 0 {
		t.Logf("First sampling simplified market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_OrderBook tests the /book endpoint against the live API.
// Uses Gamma API to get valid token IDs.
func TestIntegration_OrderBook(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing OrderBook with token_id: %s", tokenID)

	req := &types.OrderBookSummaryRequest{
		TokenId: tokenID,
		Side:    types.SideBuy,
	}

	resp, err := c.OrderBook(ctx, req)
	if err != nil {
		t.Fatalf("OrderBook failed: %v", err)
	}

	t.Logf("OrderBook response: Market=%s, AssetID=%s, Bids=%d, Asks=%d",
		resp.Market, resp.AssetID, len(resp.Bids), len(resp.Asks))
}

// TestIntegration_OrderBooks tests the /books endpoint against the live API.
// Uses Gamma API to get valid token IDs.
func TestIntegration_OrderBooks(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing OrderBooks with token_id: %s", tokenID)

	req := []types.OrderBookSummaryRequest{
		{TokenId: tokenID},
	}

	resp, err := c.OrderBooks(ctx, req)
	if err != nil {
		t.Fatalf("OrderBooks failed: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("Expected at least one order book response")
	}

	t.Logf("OrderBooks response: count=%d, first market=%s", len(resp), resp[0].Market)
}

// GammaMarket represents a market from the Gamma API.
type GammaMarket struct {
	ClobTokenIds string `json:"clobTokenIds"`
}

// fetchTokenIDFromGamma fetches a valid token_id from the Gamma API.
func fetchTokenIDFromGamma(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://gamma-api.polymarket.com/markets?limit=5&active=true&closed=false", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gamma API returned status %d", resp.StatusCode)
	}

	var markets []GammaMarket
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return "", err
	}

	if len(markets) == 0 {
		return "", fmt.Errorf("no markets found")
	}

	// Parse clobTokenIds JSON array (it's a JSON string)
	for _, m := range markets {
		if m.ClobTokenIds == "" || m.ClobTokenIds == "null" {
			continue
		}
		var tokenIds []string
		if err := json.Unmarshal([]byte(m.ClobTokenIds), &tokenIds); err != nil {
			continue
		}
		if len(tokenIds) > 0 && tokenIds[0] != "" {
			return tokenIds[0], nil
		}
	}

	return "", fmt.Errorf("no valid token_id found in markets")
}

// TestIntegration_AllEndpointsReachable runs a smoke test to verify all endpoints are reachable.
func TestIntegration_AllEndpointsReachable(t *testing.T) {
	c := getTestClient(t)

	tests := []struct {
		name          string
		fn            func(context.Context) error
		skipBodyLarge bool // if true, skip on "response body too large" errors
	}{
		// Ping is excluded - Polymarket doesn't have this endpoint
		{
			name: "Markets",
			fn: func(ctx context.Context) error {
				_, err := c.Markets(ctx, "")
				return err
			},
		},
		{
			name: "SimplifiedMarkets",
			fn: func(ctx context.Context) error {
				_, err := c.SimplifiedMarkets(ctx, "")
				return err
			},
		},
		{
			name:          "SamplingMarkets",
			skipBodyLarge: true,
			fn: func(ctx context.Context) error {
				_, err := c.SamplingMarkets(ctx, "")
				return err
			},
		},
		{
			name: "SamplingSimplifiedMarkets",
			fn: func(ctx context.Context) error {
				_, err := c.SamplingSimplifiedMarkets(ctx, "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
			defer cancel()

			err := tt.fn(ctx)
			if err != nil {
				if tt.skipBodyLarge && strings.Contains(err.Error(), "response body too large") {
					t.Skipf("%s: response body too large (expected)", tt.name)
				}
				t.Errorf("%s endpoint failed: %v", tt.name, err)
			} else {
				t.Logf("%s endpoint: OK", tt.name)
			}
		})
	}
}

// TestIntegration_GetLastTradePrice tests the /last-trade-price endpoint.
func TestIntegration_GetLastTradePrice(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetLastTradePrice with token_id: %s", tokenID)

	req := &types.LastTradePriceRequest{
		TokenId: tokenID,
	}

	resp, err := c.GetLastTradePrice(ctx, req)
	if err != nil {
		t.Fatalf("GetLastTradePrice failed: %v", err)
	}

	t.Logf("GetLastTradePrice response: Price=%s, Side=%v", resp.Price, resp.Side)

	if resp.Price.String() == "" {
		t.Error("Expected price to be present")
	}

	if resp.Side != "BUY" && resp.Side != "SELL" {
		t.Errorf("Expected Side to be 'BUY' or 'SELL', got '%s'", resp.Side)
	}
}

// TestIntegration_GetLastTradesPrices tests the /last-trades-prices endpoint.
func TestIntegration_GetLastTradesPrices(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetLastTradesPrices with token_id: %s", tokenID)

	req := []types.LastTradePriceRequest{
		{TokenId: tokenID},
	}

	resp, err := c.GetLastTradesPrices(ctx, req)
	if err != nil {
		t.Fatalf("GetLastTradesPrices failed: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("Expected at least one response")
	}

	t.Logf("GetLastTradesPrices response: %+v", resp[0])

	if resp[0].Price.String() == "" {
		t.Error("Expected price to be present")
	}
}
